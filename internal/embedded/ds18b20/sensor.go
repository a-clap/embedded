/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package ds18b20

import (
	"errors"
	"fmt"
	"github.com/a-clap/iot/internal/embedded/avg"
	"io"
	"path"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Resolution int

const (
	Resolution9Bit  Resolution = 9
	Resolution10Bit Resolution = 10
	Resolution11Bit Resolution = 11
	Resolution12Bit Resolution = 12
)

var (
	ErrUnexpectedResolution = errors.New("unexpected resolution")
)

type Readings struct {
	ID          string    `json:"id"`
	Temperature float32   `json:"temperature"`
	Average     float32   `json:"average"`
	Stamp       time.Time `json:"stamp"`
	Error       string    `json:"error"`
}

type Sensor struct {
	FileOpener
	id, bus                         string
	temperaturePath, resolutionPath string
	polling                         atomic.Bool
	fin, stop                       chan struct{}
	data                            chan Readings
	average                         *avg.Avg[float32]
	cfg                             SensorConfig
	readings                        []Readings
	mtx                             *sync.Mutex
}

type SensorConfig struct {
	ID           string        `json:"id"`
	Correction   float32       `json:"correction"`
	Resolution   Resolution    `json:"resolution"`
	PollInterval time.Duration `json:"poll_interval"`
	Samples      uint          `json:"samples"`
}

func NewSensor(o FileOpener, id, basePath string) (*Sensor, error) {
	bus := basePath[strings.LastIndex(basePath, "/")+1:]

	s := &Sensor{
		FileOpener:      o,
		id:              id,
		bus:             bus,
		temperaturePath: path.Join(basePath, id, "temperature"),
		resolutionPath:  path.Join(basePath, id, "resolution"),
		polling:         atomic.Bool{},
		mtx:             &sync.Mutex{},
		cfg: SensorConfig{
			ID:           id,
			Correction:   0,
			Resolution:   Resolution11Bit,
			PollInterval: 100 * time.Millisecond,
			Samples:      10,
		},
	}
	var err error
	if s.average, err = avg.New[float32](s.cfg.Samples); err != nil {
		return nil, err
	}

	if s.cfg.Resolution, err = s.resolution(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Sensor) Name() (bus string, id string) {
	return s.bus, s.id
}

func (s *Sensor) Poll() (err error) {
	if s.polling.Load() {
		return ErrAlreadyPolling
	}

	s.polling.Store(true)
	s.fin = make(chan struct{})
	s.stop = make(chan struct{})
	s.data = make(chan Readings, 10)
	go s.poll()

	return nil
}

func (s *Sensor) Temperature() (actual, avg float32, err error) {
	avg = s.Average()

	conv, err := s.readFile(s.temperaturePath)
	if err != nil {
		return
	}
	length := len(conv)
	if length > 3 {
		conv = conv[:length-3] + "." + conv[length-3:]
	} else {
		leading := "0."
		for length < 3 {
			leading += "0"
			length++
		}
		conv = leading + conv
	}
	t64, err := strconv.ParseFloat(conv, 32)
	if err != nil {
		return
	}
	tmp := float32(t64) + s.cfg.Correction
	s.average.Add(tmp)
	return tmp, s.Average(), nil
}

func (s *Sensor) Average() float32 {
	return s.average.Average()
}

func (s *Sensor) GetReadings() []Readings {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if len(s.readings) == 0 {
		return nil
	}
	c := make([]Readings, len(s.readings))
	copy(c, s.readings)
	s.readings = nil
	return c
}
func (s *Sensor) add(r Readings) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.readings = append(s.readings, r)
	if len(s.readings) > 100 {
		s.readings = s.readings[1:]
	}
}
func (s *Sensor) Configure(config SensorConfig) error {
	if s.cfg.Samples != config.Samples {
		if err := s.average.Resize(config.Samples); err != nil {
			return err
		}
		s.cfg.Samples = config.Samples
	}

	if s.cfg.Resolution != config.Resolution {
		if err := s.setResolution(config.Resolution); err != nil {
			return err
		}
		s.cfg.Resolution = config.Resolution
	}

	s.cfg.PollInterval = config.PollInterval
	s.cfg.Correction = config.Correction
	return nil
}

func (s *Sensor) GetConfig() SensorConfig {
	return s.cfg
}

func (s *Sensor) Close() error {
	if !s.polling.Load() {
		// Nothing to do
		return nil
	}

	// Close stop channel to signal finish of polling
	close(s.stop)
	// Unblock poll
	for range s.data {
	}
	// Wait until finish
	for range s.fin {
	}

	return nil
}

func (s *Sensor) resolution() (r Resolution, err error) {
	res, err := s.readFile(s.resolutionPath)
	if err != nil {
		return r, err
	}

	maybeRes, err := strconv.ParseInt(res, 10, 32)
	if err != nil {
		Log.Error(err)
		return
	}
	r = Resolution(maybeRes)
	if r < Resolution9Bit || r > Resolution12Bit {
		return r, fmt.Errorf("%w : %v", ErrUnexpectedResolution, res)
	}

	return
}

func (s *Sensor) setResolution(res Resolution) error {
	resFile, err := s.Open(s.resolutionPath)
	if err != nil {
		Log.Error(err)
		return err
	}

	defer func() {
		err = resFile.Close()
		if err != nil {
			Log.Error(err)
		}
	}()

	buf := strconv.FormatInt(int64(res), 10) + "\r\n"
	if _, err := resFile.Write([]byte(buf)); err != nil {
		return err
	}
	return nil
}

func (s *Sensor) poll() {
	go func() {
		for data := range s.data {
			s.add(data)
		}
	}()

	for s.polling.Load() {
		select {
		case <-s.stop:
			s.polling.Store(false)
		case <-time.After(s.cfg.PollInterval):
			actual, average, err := s.Temperature()
			e := ""
			if err != nil {
				e = err.Error()
			}
			s.data <- Readings{
				ID:          s.id,
				Temperature: actual,
				Average:     average,
				Stamp:       time.Now(),
				Error:       e,
			}
		}
	}
	close(s.data)
	// For sure there won't be more data
	// Sensor created channel (and is the sender side), so should close
	close(s.fin)
}

func (s *Sensor) readFile(path string) (string, error) {
	f, err := s.Open(path)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := f.Close(); err != nil {
			Log.Error(err)
		}
	}()

	// Files used by ds will have just few bytes, io.ReadAll seems okay for that purpose
	buf, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}

	return strings.TrimRight(string(buf), "\r\n"), nil
}

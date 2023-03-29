/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package ds18b20

import (
	"errors"
	"fmt"
	"io"
	"path"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/a-clap/iot/pkg/avg"
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
	Temperature float64   `json:"temperature"`
	Average     float64   `json:"average"`
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
	average                         *avg.Avg
	cfg                             SensorConfig
	readings                        []Readings
	mtx                             *sync.Mutex
}

type SensorConfig struct {
	ID           string        `json:"id"`
	Correction   float64       `json:"correction"`
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
	if s.average, err = avg.New(s.cfg.Samples); err != nil {
		return nil, &Error{Bus: s.bus, ID: s.id, Op: "NewSensor", Err: err.Error()}
	}

	if s.cfg.Resolution, err = s.resolution(); err != nil {
		return nil, &Error{Bus: s.bus, ID: s.id, Op: "NewSensor.resolution", Err: err.Error()}
	}

	return s, nil
}

func (s *Sensor) Name() (bus string, id string) {
	return s.bus, s.id
}

func (s *Sensor) Poll() (err error) {
	if s.polling.Load() {
		return &Error{Bus: s.bus, ID: s.id, Op: "Poll", Err: ErrAlreadyPolling.Error()}
	}

	s.polling.Store(true)
	s.fin = make(chan struct{})
	s.stop = make(chan struct{})
	s.data = make(chan Readings, 10)
	go s.poll()

	return nil
}

func (s *Sensor) Temperature() (actual, avg float64, err error) {
	conv, err := s.readFile(s.temperaturePath)
	if err != nil {
		err = &Error{
			Bus: s.bus,
			ID:  s.id,
			Op:  "readFile :" + s.temperaturePath,
			Err: err.Error(),
		}
		return 0, 0, err
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
	t64, err := strconv.ParseFloat(conv, 64)
	if err != nil {
		err = &Error{
			Bus: s.bus,
			ID:  s.id,
			Op:  "ParseFloat :" + conv,
			Err: err.Error(),
		}
		return 0, 0, err
	}
	tmp := t64 + s.cfg.Correction
	s.average.Add(tmp)
	return tmp, s.Average(), nil
}

func (s *Sensor) Average() float64 {
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
			return &Error{
				Bus: s.bus,
				ID:  s.id,
				Op:  "Configure.Resize",
				Err: err.Error(),
			}
		}
		s.cfg.Samples = config.Samples
	}

	if s.cfg.Resolution != config.Resolution {
		if err := s.setResolution(config.Resolution); err != nil {
			return &Error{
				Bus: s.bus,
				ID:  s.id,
				Op:  "Configure.setResolution",
				Err: err.Error(),
			}
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
		return
	}

	maybeRes, err := strconv.ParseInt(res, 10, 32)
	if err != nil {
		return
	}
	r = Resolution(maybeRes)
	if r < Resolution9Bit || r > Resolution12Bit {
		return r, fmt.Errorf("%w : %v", ErrUnexpectedResolution, res)
	}

	return
}

func (s *Sensor) setResolution(res Resolution) (err error) {
	resFile, err := s.Open(s.resolutionPath)
	if err != nil {
		return
	}

	defer func() {
		errOnClose := resFile.Close()
		// If there is already error, we don't want to hide it
		if err == nil {
			err = errOnClose
		}
	}()

	buf := strconv.FormatInt(int64(res), 10) + "\r\n"
	if _, err = resFile.Write([]byte(buf)); err != nil {
		return
	}
	return
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
	// For sure there won't be more data
	close(s.data)
	// Notify we are done
	close(s.fin)
}

func (s *Sensor) readFile(path string) (r string, err error) {
	f, err := s.Open(path)
	if err != nil {
		return "", err
	}

	defer func() {
		errOnClose := f.Close()
		// If there is already error, we don't want to hide it
		if err == nil {
			err = errOnClose
		}
	}()

	// Files used by ds will have just few bytes, io.ReadAll seems okay for that purpose
	buf, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}

	return strings.TrimRight(string(buf), "\r\n"), err
}

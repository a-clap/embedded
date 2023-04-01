/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package ds18b20

import (
	"errors"
	"fmt"
	"path"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/a-clap/iot/pkg/avg"
)

type Resolution int

// Possible resolutions
const (
	Resolution9Bit  Resolution = 9
	Resolution10Bit Resolution = 10
	Resolution11Bit Resolution = 11
	Resolution12Bit Resolution = 12
)

var (
	ErrUnexpectedResolution = errors.New("unexpected resolution")
)

// Readings are returned, when Sensor is used in Poll mode
type Readings struct {
	ID          string    `json:"id"`
	Temperature float64   `json:"temperature"`
	Average     float64   `json:"average"`
	Stamp       time.Time `json:"stamp"`
	Error       string    `json:"error"`
}

// Sensor represents DS18b20
type Sensor struct {
	FileReaderWriter
	fullID                          string
	id                              string
	temperaturePath, resolutionPath string
	polling                         atomic.Bool
	fin, stop                       chan struct{}
	data                            chan Readings
	average                         *avg.Avg
	cfg                             SensorConfig
	readings                        []Readings
	mtx                             *sync.Mutex
}

// SensorConfig allows user to configure Sensor (except ID, which is unique and can't be changed)
type SensorConfig struct {
	Name         string        `json:"name"`
	ID           string        `json:"id"`
	Correction   float64       `json:"correction"`
	Resolution   Resolution    `json:"resolution"`
	PollInterval time.Duration `json:"poll_interval"`
	Samples      uint          `json:"samples"`
}

// NewSensor creates new sensor based on args
func NewSensor(o FileReaderWriter, id, basePath string) (*Sensor, error) {
	bus := basePath[strings.LastIndex(basePath, "/")+1:]

	s := &Sensor{
		FileReaderWriter: o,
		id:               id,
		fullID:           bus + ":" + id,
		temperaturePath:  path.Join(basePath, id, "temperature"),
		resolutionPath:   path.Join(basePath, id, "resolution"),
		polling:          atomic.Bool{},
		fin:              nil,
		stop:             nil,
		data:             nil,
		average:          nil,
		cfg: SensorConfig{
			ID:           id,
			Correction:   0,
			Resolution:   Resolution11Bit,
			PollInterval: 100 * time.Millisecond,
			Samples:      10,
		},
		readings: nil,
		mtx:      &sync.Mutex{},
	}
	s.average = avg.New(s.cfg.Samples)

	var err error
	if s.cfg.Resolution, err = s.resolution(); err != nil {
		return nil, fmt.Errorf("NewSensor.resolution {ID: %v}: %w", s.fullID, err)
	}

	return s, nil
}

// Name returns user provided name
func (s *Sensor) Name() string {
	return s.cfg.Name
}

// ID returns Sensor hardware id in id
func (s *Sensor) ID() string {
	return s.id
}

// FullID returns Sensor hardware id in convention: w1Path:id
func (s *Sensor) FullID() string {
	return s.fullID
}

// Poll is an option to run temperature updates in background
// After calling Poll, user can get data from GetReadings
func (s *Sensor) Poll() {
	if s.polling.Load() {
		return
	}

	s.polling.Store(true)
	s.fin = make(chan struct{})
	s.stop = make(chan struct{})
	s.data = make(chan Readings, 10)
	go s.poll()
}

// Temperature returns current temperature and average (which is based on Samples)
func (s *Sensor) Temperature() (actual, avg float64, err error) {
	conv, err := s.readFile(s.temperaturePath)
	if err != nil {
		err = fmt.Errorf("Temperature.readFile {ID: %v, path: %v}: %w", s.fullID, s.temperaturePath, err)
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
		err = fmt.Errorf("Temperature.ParseFloat {ID: %v, path: %v, value:%v}: %w", s.fullID, s.temperaturePath, conv, err)
		return 0, 0, err
	}
	tmp := t64 + s.cfg.Correction
	s.average.Add(tmp)
	return tmp, s.Average(), nil
}

// Average returns current average temperature
func (s *Sensor) Average() float64 {
	return s.average.Average()
}

// GetReadings returns all collected readings and then clears data
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

// Configure allows user to configure sensor with SensorConfig
func (s *Sensor) Configure(config SensorConfig) error {
	s.cfg.Name = config.Name

	if s.cfg.Samples != config.Samples {
		s.average.Resize(config.Samples)
		s.cfg.Samples = config.Samples
	}

	if s.cfg.Resolution != config.Resolution {
		if err := s.setResolution(config.Resolution); err != nil {
			return fmt.Errorf("Configure.setResolution {ID: %v, Resolution: %v}: %w", s.fullID, config.Resolution, err)
		}
		s.cfg.Resolution = config.Resolution
	}

	s.cfg.PollInterval = config.PollInterval
	s.cfg.Correction = config.Correction
	return nil
}

// GetConfig returns current config
func (s *Sensor) GetConfig() SensorConfig {
	return s.cfg
}

// Close should be called, if user used Poll
func (s *Sensor) Close() {
	if !s.polling.Load() {
		// Nothing to do
		return
	}

	// Close stop channel to signal finish of polling
	close(s.stop)
	// Unblock poll
	for range s.data {
	}
	// Wait until finish
	for range s.fin {
	}

	return
}

func (s *Sensor) resolution() (r Resolution, err error) {
	res, err := s.readFile(s.resolutionPath)
	if err != nil {
		return r, fmt.Errorf("resolution: {ID: %v}: %w", s.fullID, err)
	}

	maybeRes, err := strconv.ParseInt(res, 10, 32)
	if err != nil {
		return r, fmt.Errorf("resolution: {ID: %v, Resolution: %v}: %w", s.fullID, res, err)
	}
	r = Resolution(maybeRes)
	if r < Resolution9Bit || r > Resolution12Bit {
		return r, fmt.Errorf("resolution: {ID: %v, Resolution: %v}: %w", s.fullID, res, ErrUnexpectedResolution)
	}

	return
}

func (s *Sensor) setResolution(res Resolution) (err error) {
	buf := strconv.FormatInt(int64(res), 10) + "\r\n"
	if err = s.WriteFile(s.resolutionPath, []byte(buf)); err != nil {
		return fmt.Errorf("setResolution {Path: %v}: %w", s.resolutionPath, err)
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
	buf, err := s.ReadFile(path)
	if err != nil {
		return
	}
	return strings.TrimRight(string(buf), "\r\n"), err
}

func (s *Sensor) add(r Readings) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.readings = append(s.readings, r)
	if len(s.readings) > 100 {
		s.readings = s.readings[1:]
	}
}

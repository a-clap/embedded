/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package max31865

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/a-clap/embedded/pkg/avg"
)

// Sensor is representation of each Max31865
type Sensor struct {
	ReadWriteCloser
	configReg       *configReg
	r               *rtd
	trig, fin, stop chan struct{}
	err             chan error
	data            chan Readings
	cfg             SensorConfig
	average         *avg.Avg
	polling         atomic.Bool
	ready           Ready
	readings        []Readings
	mtx             sync.Mutex
}

// SensorConfig holds configuration for Sensor
type SensorConfig struct {
	Name         string        `json:"name"`
	ID           string        `json:"id"`
	Correction   float64       `json:"correction"`
	ASyncPoll    bool          `json:"a_sync_poll"`
	PollInterval time.Duration `json:"poll_interval"`
	Samples      uint          `json:"samples"`
}

// Readings is a structure returned, when user uses Poll
type Readings struct {
	ID          string    `json:"id"`
	Temperature float64   `json:"temperature"`
	Average     float64   `json:"average"`
	Stamp       time.Time `json:"stamp"`
	Error       string    `json:"error"`
}

// NewSensor creates Sensor with provided options
func NewSensor(options ...Option) (*Sensor, error) {
	s := &Sensor{
		ReadWriteCloser: nil,
		r:               &rtd{},
		configReg:       newConfig(),
		readings:        nil,
		mtx:             sync.Mutex{},
		cfg: SensorConfig{
			ID:           "",
			Correction:   0,
			ASyncPoll:    false,
			PollInterval: 33 * time.Millisecond,
			Samples:      10,
		},
	}
	s.average = avg.New(s.cfg.Samples)

	for _, opt := range options {
		if err := opt(s); err != nil {
			return nil, fmt.Errorf("NewSensor: %w", err)
		}
	}

	// verify after parsing opts
	if err := s.verify(); err != nil {
		return nil, fmt.Errorf("NewSensor.Verify: %w", err)
	}

	// Do initial regConfig
	if err := s.config(); err != nil {
		return nil, fmt.Errorf("NewSensor.config: %w", err)
	}

	return s, nil
}

// Poll allows user to enable background temperature updates
// Then data can be retrieved by calling GetReadings()
func (s *Sensor) Poll() (err error) {
	if s.polling.Load() {
		return fmt.Errorf("Poll {ID: %v}: %w", s.ID(), ErrAlreadyPolling)
	}

	s.fin = make(chan struct{})
	s.stop = make(chan struct{})
	s.err = make(chan error, 10)
	s.data = make(chan Readings, 10)

	s.polling.Store(true)
	if s.cfg.ASyncPoll {
		err = s.prepareAsyncPoll()
	} else {
		err = s.prepareSyncPoll(s.cfg.PollInterval)
	}

	if err != nil {
		s.polling.Store(false)
		return fmt.Errorf("Poll {ID: %v}: %w", s.ID(), err)
	}
	go s.poll()

	return
}

// GetReadings returns accumulated data by Poll, clears data on read
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

// Configure is a way to set Config
func (s *Sensor) Configure(config SensorConfig) error {
	if config.ASyncPoll {
		if s.ready == nil {
			return fmt.Errorf("Configure {ID: %v, Config: %v}: %w", s.ID(), config, ErrNoReadyInterface)
		}
	}
	if s.cfg.Samples != config.Samples {
		s.average.Resize(config.Samples)
		s.cfg.Samples = config.Samples
	}
	s.cfg.Name = config.Name
	s.cfg.ASyncPoll = config.ASyncPoll
	s.cfg.PollInterval = config.PollInterval
	s.cfg.Correction = config.Correction

	return nil
}

// GetConfig returns current Config
func (s *Sensor) GetConfig() SensorConfig {
	return s.cfg
}

// Average returns average temperature
func (s *Sensor) Average() float64 {
	return s.average.Average()
}

// Temperature returns actual temperature and average
func (s *Sensor) Temperature() (actual float64, average float64, err error) {
	r, err := s.read(regConf, regFault+1)
	if err != nil {
		//	can't do much about it
		err = fmt.Errorf("Temperature.read {ID: %v}: %w", s.ID(), err)
		return
	}
	err = s.r.update(r[regRtdMsb], r[regRtdLsb])
	if err != nil {
		// Not handling error here, should have happened on previous call
		_ = s.clearFaults()
		// make error more specific
		err = fmt.Errorf("Temperature.rtd {ID: %v, Reg: %v, Cause: %v}: %w", s.ID(), r[regFault], errorCauses(r[regFault], s.configReg.wiring), err)
		return
	}
	tmp := s.r.toTemperature(s.configReg.refRes, s.configReg.nominalRes)
	s.average.Add(tmp + s.cfg.Correction)
	return tmp, s.average.Average(), nil
}

// Close should be always called, if user used Poll
func (s *Sensor) Close() error {
	if !s.polling.Load() {
		return nil
	}
	// Close stop channel, notify poll
	close(s.stop)
	// Unblock poll
	for range s.data {
	}
	// Wait until finish
	for range s.fin {
	}

	return nil
}

// ID returns unique ID
func (s *Sensor) ID() string {
	return s.cfg.ID
}

func (s *Sensor) prepareSyncPoll(pollTime time.Duration) error {
	s.trig = make(chan struct{})
	go func(s *Sensor, pollTime time.Duration) {
		for s.polling.Load() {
			<-time.After(pollTime)
			s.callback()
		}
		close(s.trig)
	}(s, pollTime)

	return nil
}

func (s *Sensor) prepareAsyncPoll() error {
	if s.ready == nil {
		return ErrNoReadyInterface
	}
	s.trig = make(chan struct{}, 1)
	return s.ready.Open(s.callback)
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
		case <-s.trig:
			tmp, average, err := s.Temperature()
			var e string
			if err != nil {
				e = err.Error()
			}

			s.data <- Readings{
				ID:          s.ID(),
				Temperature: tmp,
				Average:     average,
				Stamp:       time.Now(),
				Error:       e,
			}
		case e := <-s.err:
			s.data <- Readings{
				Error: e.Error(),
			}
		}
	}
	// For sure there won't be more data
	close(s.data)
	if s.cfg.ASyncPoll {
		s.ready.Close()
		close(s.trig)
	}

	// Notify user that we are done
	s.fin <- struct{}{}
	close(s.fin)
}

func (s *Sensor) clearFaults() error {
	return s.write(regConf, []byte{s.configReg.clearFaults()})
}

func (s *Sensor) config() error {
	err := s.write(regConf, []byte{s.configReg.reg()})
	return err
}

func (s *Sensor) read(addr byte, len int) ([]byte, error) {
	// We need to create slice with 1 byte more
	w := make([]byte, len+1)
	w[0] = addr
	r, err := s.ReadWrite(w)
	if err != nil {
		return nil, err
	}
	// First byte is useless
	return r[1:], nil
}

func (s *Sensor) write(addr byte, w []byte) error {
	buf := []byte{addr | 0x80}
	buf = append(buf, w...)
	_, err := s.ReadWrite(buf)
	return err
}

func (s *Sensor) verify() error {
	// Check if interface exists
	if s.ReadWriteCloser == nil {
		return ErrNoReadWriteCloser
	}
	// Check interface itself
	const size = regFault + 2
	buf := make([]byte, size)
	buf[0] = regConf
	r, err := s.ReadWrite(buf)
	if err != nil {
		return ErrInterface
	}
	checkReadings := func(expected byte) bool {
		for _, elem := range r {
			if elem != expected {
				return false
			}
		}
		return true
	}

	if onlyZeroes := checkReadings(0); onlyZeroes {
		return ErrReadZeroes
	}

	if onlyFF := checkReadings(0xff); onlyFF {
		return ErrReadFF
	}

	return nil
}

func (s *Sensor) callback() {
	// We don't want to block on channel write, as it may be isr
	select {
	case s.trig <- struct{}{}:
		return
	default:
		s.err <- ErrTooMuchTriggers
	}
}

func (s *Sensor) add(r Readings) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.readings = append(s.readings, r)
	if len(s.readings) > 100 {
		s.readings = s.readings[1:]
	}
}

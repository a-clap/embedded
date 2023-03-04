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

	"github.com/a-clap/iot/internal/embedded/avg"
)

type Error struct {
	ID  string `json:"ID"`
	Op  string `json:"op"`
	Err string `json:"error"`
}

func (e *Error) Error() string {
	if e.Err == "" {
		return "<nil>"
	}
	s := e.Op
	if e.ID != "" {
		s += ":" + e.ID
	}
	s += ": " + e.Err
	return s
}

type Sensor struct {
	ReadWriteCloser
	configReg       *configReg
	r               *rtd
	trig, fin, stop chan struct{}
	data            chan Readings
	cfg             SensorConfig
	average         *avg.Avg
	polling         atomic.Bool
	ready           Ready
	readings        []Readings
	mtx             sync.Mutex
}

type SensorConfig struct {
	ID           string        `json:"id"`
	Correction   float64       `json:"correction"`
	ASyncPoll    bool          `json:"a_sync_poll"`
	PollInterval time.Duration `json:"poll_interval"`
	Samples      uint          `json:"samples"`
}

type Readings struct {
	ID          string    `json:"id"`
	Temperature float64   `json:"temperature"`
	Average     float64   `json:"average"`
	Stamp       time.Time `json:"stamp"`
	Error       string    `json:"error"`
}

func NewSensor(options ...Option) (*Sensor, error) {
	s := &Sensor{
		ReadWriteCloser: nil,
		r:               newRtd(),
		configReg:       newConfig(),
		readings:        nil,
		mtx:             sync.Mutex{},
		cfg: SensorConfig{
			ID:           "",
			Correction:   0,
			ASyncPoll:    false,
			PollInterval: 100 * time.Millisecond,
			Samples:      10,
		},
	}
	for _, opt := range options {
		if err := opt(s); err != nil {
			return nil, &Error{Op: "Max.NewSensor", Err: err.Error()}
		}
	}

	var err error
	s.average, err = avg.New(s.cfg.Samples)
	if err != nil {
		return nil, &Error{ID: s.ID(), Op: "Max.NewSensor.Avg", Err: err.Error()}
	}
	// verify after parsing opts
	if err := s.verify(); err != nil {
		return nil, &Error{ID: s.ID(), Op: "Max.NewSensor.Verify", Err: err.Error()}
	}

	// Do initial regConfig
	if err := s.config(); err != nil {
		return nil, &Error{ID: s.ID(), Op: "Max.NewSensor.config", Err: err.Error()}
	}

	return s, nil
}

func (s *Sensor) Poll() (err error) {
	if s.polling.Load() {
		return &Error{ID: s.ID(), Op: "Poll", Err: ErrAlreadyPolling.Error()}
	}

	s.polling.Store(true)
	if s.cfg.ASyncPoll {
		err = s.prepareAsyncPoll()
	} else {
		err = s.prepareSyncPoll(s.cfg.PollInterval)
	}

	if err != nil {
		s.polling.Store(false)
		return &Error{ID: s.ID(), Op: "Poll", Err: err.Error()}
	}

	s.fin = make(chan struct{})
	s.stop = make(chan struct{})
	s.data = make(chan Readings, 10)
	go s.poll()

	return nil
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
			return &Error{ID: s.ID(), Op: "Configure.Avg.Resize", Err: err.Error()}
		}
		s.cfg.Samples = config.Samples
	}
	if config.ASyncPoll {
		if s.ready == nil {
			return &Error{ID: s.ID(), Op: "Configure", Err: ErrNoReadyInterface.Error()}
		}
	}
	s.cfg.ASyncPoll = config.ASyncPoll

	s.cfg.PollInterval = config.PollInterval
	s.cfg.Correction = config.Correction
	return nil
}
func (s *Sensor) GetConfig() SensorConfig {
	return s.cfg
}

func (s *Sensor) prepareSyncPoll(pollTime time.Duration) error {
	s.trig = make(chan struct{})
	go func(s *Sensor, pollTime time.Duration) {
		for s.polling.Load() {
			<-time.After(pollTime)
			if s.polling.Load() {
				s.trig <- struct{}{}
			}
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
	return s.ready.Open(callback, s)
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
			e := ""
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

func (s *Sensor) Average() float64 {
	return s.average.Average()
}

func (s *Sensor) Temperature() (actual float64, average float64, err error) {
	r, err := s.read(regConf, regFault+1)
	if err != nil {
		//	can't do much about it
		err = &Error{ID: s.ID(), Op: "Temperature.read", Err: err.Error()}
		return
	}
	err = s.r.update(r[regRtdMsb], r[regRtdLsb])
	if err != nil {
		// Not handling error here, should have happened on previous call
		_ = s.clearFaults()
		// make error more specific
		err = fmt.Errorf("%w: errorReg: %v, posibble causes: %v", err, r[regFault], errorCauses(r[regFault], s.configReg.wiring))
		err = &Error{ID: s.ID(), Op: "Temperature.update", Err: err.Error()}
		return
	}
	tmp := s.r.toTemperature(s.configReg.refRes, s.configReg.rNominal)
	s.average.Add(tmp)
	return tmp, s.average.Average(), nil
}

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

	return s.ReadWriteCloser.Close()
}

func (s *Sensor) ID() string {
	return s.configReg.id
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

func callback(args any) error {
	s, ok := args.(*Sensor)
	if !ok {
		return ErrWrongArgs
	}
	// We don't want to block on channel write, as it may be isr
	select {
	case s.trig <- struct{}{}:
		return nil
	default:
		return ErrTooMuchTriggers
	}
}

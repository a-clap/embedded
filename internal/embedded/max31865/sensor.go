package max31865

import (
	"fmt"
	"strconv"
	"time"
)

type Sensor struct {
	ReadWriteCloser
	cfg  *config
	r    *rtd
	trig chan struct{}
	fin  chan struct{}
	stop chan struct{}
	data chan Readings
}

func (s *Sensor) Poll(data chan Readings, pollTime time.Duration) (err error) {
	if s.cfg.polling.Load() {
		return ErrAlreadyPolling
	}

	s.cfg.polling.Store(true)
	if pollTime == -1 {
		err = s.prepareAsyncPoll()
	} else {
		err = s.prepareSyncPoll(pollTime)
	}

	if err != nil {
		s.cfg.polling.Store(false)
		return err
	}

	s.fin = make(chan struct{})
	s.stop = make(chan struct{})
	s.data = data
	go s.poll()

	return nil
}

func (s *Sensor) prepareSyncPoll(pollTime time.Duration) error {
	s.trig = make(chan struct{})
	go func(s *Sensor, pollTime time.Duration) {
		for s.cfg.polling.Load() {
			<-time.After(pollTime)
			if s.cfg.polling.Load() {
				s.trig <- struct{}{}
			}
		}
		close(s.trig)
	}(s, pollTime)

	return nil
}

func (s *Sensor) prepareAsyncPoll() error {
	if s.cfg.ready == nil {
		return ErrNoReadyInterface
	}
	s.trig = make(chan struct{}, 1)
	return s.cfg.ready.Open(callback, s)
}

func (s *Sensor) poll() {
	for s.cfg.polling.Load() {
		select {
		case <-s.stop:
			s.cfg.polling.Store(false)
		case <-s.trig:
			tmp, err := s.Temperature()
			s.data <- Readings{
				ID:          s.ID(),
				Temperature: tmp,
				Stamp:       time.Now(),
				Error:       err,
			}
		}
	}
	// For sure there won't be more data
	close(s.data)
	if s.cfg.pollType == async {
		s.cfg.ready.Close()
		close(s.trig)
	}

	// Notify user that we are done
	s.fin <- struct{}{}
	close(s.fin)
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

func (s *Sensor) Temperature() (string, error) {
	r, err := s.read(regConf, regFault+1)
	if err != nil {
		//	can't do much about it
		return "", err
	}
	err = s.r.update(r[regRtdMsb], r[regRtdLsb])
	if err != nil {
		// Not handling error here, should have happened on previous call
		_ = s.clearFaults()
		// make error more specific
		err = fmt.Errorf("%w: errorReg: %v, posibble causes: %v", err, r[regFault], errorCauses(r[regFault], s.cfg.wiring))
		return "", err
	}
	rtd := s.r.rtd()
	tmp := rtdToTemperature(rtd, s.cfg.refRes, s.cfg.rNominal)

	return strconv.FormatFloat(float64(tmp), 'f', -1, 32), nil
}

func (s *Sensor) Close() error {
	if s.cfg.polling.Load() {
		s.stop <- struct{}{}
		// Close stop channel, not needed anymore
		close(s.stop)
		// Unblock poll
		for range s.data {
		}
		// Wait until finish
		for range s.fin {
		}
	}

	return s.ReadWriteCloser.Close()
}

func (s *Sensor) ID() string {
	return string(s.cfg.id)
}

func newSensor(options ...Option) (*Sensor, error) {
	s := &Sensor{
		ReadWriteCloser: nil,
		r:               newRtd(),
		cfg:             newConfig(),
	}
	for _, opt := range options {
		if err := opt(s); err != nil {
			return nil, err
		}
	}
	// verify after parsing opts
	if err := s.verify(); err != nil {
		return nil, err
	}

	// Do initial regConfig
	if err := s.config(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Sensor) clearFaults() error {
	return s.write(regConf, []byte{s.cfg.clearFaults()})
}

func (s *Sensor) config() error {
	err := s.write(regConf, []byte{s.cfg.reg()})
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

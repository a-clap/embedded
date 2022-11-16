package max31865

import (
	"fmt"
	"io"
	"strconv"
	"time"
)

type Sensor interface {
	io.Closer
	ID() string
	Temperature() (float32, error)
	Poll(data chan Readings, pollTime time.Duration) (err error)
}

var _ Sensor = &sensor{}
var _ Readings = &readings{}

type sensor struct {
	Transfer
	cfg    config
	regCfg *regConfig
	r      *rtd
	trig   chan struct{}
	fin    chan struct{}
	stop   chan struct{}
	data   chan Readings
}

type Readings interface {
	ID() string
	Get() (temperature string, timestamp time.Time, err error)
}

type readings struct {
	id, temperature string
	timestamp       time.Time
	err             error
}

func (s *sensor) Poll(data chan Readings, pollTime time.Duration) (err error) {
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

func (s *sensor) prepareSyncPoll(pollTime time.Duration) error {
	s.trig = make(chan struct{})
	go func(s *sensor, pollTime time.Duration) {
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

func (s *sensor) prepareAsyncPoll() error {
	if s.cfg.ready == nil {
		return ErrNoReadyInterface
	}
	s.trig = make(chan struct{}, 1)
	return s.cfg.ready.Open(callback, s)
}

func (s *sensor) poll() {
	for s.cfg.polling.Load() {
		select {
		case <-s.stop:
			s.cfg.polling.Store(false)
		case <-s.trig:
			tmp, err := s.Temperature()
			r := readings{
				id:  s.ID(),
				err: nil,
			}
			if err != nil {
				r.err = err
			} else {
				r.temperature = strconv.FormatFloat(float64(tmp), 'f', -1, 32)
				r.timestamp = time.Now()
			}
			s.data <- r
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
	s, ok := args.(*sensor)
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

func (s *sensor) Temperature() (tmp float32, err error) {
	r, err := s.read(regConf, regFault+1)
	if err != nil {
		//	can't do much about it
		return
	}
	err = s.r.update(r[regRtdMsb], r[regRtdLsb])
	if err != nil {
		// Not handling error here, should have happened on previous call
		_ = s.clearFaults()
		// make error more specific
		err = fmt.Errorf("%w: errorReg: %v, posibble causes: %v", err, r[regFault], errorCauses(r[regFault], s.cfg.wiring))
		return
	}
	rtd := s.r.rtd()
	return rtdToTemperature(rtd, s.cfg.refRes, s.cfg.rNominal), nil
}

func (s *sensor) Close() error {
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

	return s.Transfer.Close()
}

func (s *sensor) ID() string {
	return string(s.cfg.id)
}

func newSensor(t Transfer, args ...any) (*sensor, error) {
	s := &sensor{
		Transfer: t,
		regCfg:   newRegConfig(),
		r:        newRtd(),
		cfg:      newConfig(),
	}

	s.parse(args...)
	// Do initial regConfig
	err := s.config()
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *sensor) parse(args ...any) {
	for _, arg := range args {
		switch arg := arg.(type) {
		case Ready:
			s.cfg.ready = arg
		case ID:
			s.cfg.id = arg
		case Wiring:
			s.cfg.wiring = arg
			s.regCfg.setWiring(arg)
		case RefRes:
			s.cfg.refRes = arg
		case RNominal:
			s.cfg.rNominal = arg
		}
	}
}

func (s *sensor) clearFaults() error {
	return s.write(regConf, []byte{s.regCfg.clearFaults()})
}

func (s *sensor) config() error {
	err := s.write(regConf, []byte{s.regCfg.reg()})
	return err
}

func (s *sensor) read(addr byte, len int) ([]byte, error) {
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

func (s *sensor) write(addr byte, w []byte) error {
	buf := []byte{addr | 0x80}
	buf = append(buf, w...)
	_, err := s.ReadWrite(buf)
	return err
}

func (r readings) ID() string {
	return r.id
}

func (r readings) Get() (temperature string, timestamp time.Time, err error) {
	return r.temperature, r.timestamp, r.err
}

package max31865

import (
	"fmt"
	"strconv"
	"time"
)

type Sensor interface {
	ID() string
	Temperature() (float32, error)
	Poll(data chan<- Readings, stopCh <-chan struct{}, pollTime time.Duration) (finCh <-chan struct{}, errCh <-chan error, err error)
}

var _ Sensor = &sensor{}
var _ Readings = &readings{}

type sensor struct {
	Transfer
	cfg    config
	regCfg *regConfig
	r      *rtd
	trig   chan struct{}
}

type Readings interface {
	Get() (id string, temperature string, timestamp time.Time)
}

type readings struct {
	id, temperature string
	timestamp       time.Time
}

func (s *sensor) Poll(data chan<- Readings, stopCh <-chan struct{}, pollTime time.Duration) (finCh <-chan struct{}, errCh <-chan error, err error) {
	if s.cfg.polling {
		return nil, nil, ErrAlreadyPolling
	}

	s.cfg.polling = true
	if pollTime == -1 {
		err = s.prepareAsyncPoll()
	} else {
		err = s.prepareSyncPoll(pollTime)
	}

	if err != nil {
		s.cfg.polling = false
		return nil, nil, err
	}

	finChan := make(chan struct{})
	errChan := make(chan error)

	go s.poll(data, stopCh, finChan, errChan)

	return finChan, errChan, nil
}

func (s *sensor) prepareSyncPoll(pollTime time.Duration) error {
	s.trig = make(chan struct{})
	go func(s *sensor, pollTime time.Duration) {
		for s.cfg.polling {
			<-time.After(pollTime)
			if s.cfg.polling {
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
	return s.cfg.ready.Open(callback, s)
}

func (s *sensor) poll(data chan<- Readings, stopCh <-chan struct{}, finCh chan struct{}, errCh chan error) {
	for s.cfg.polling {
		select {
		case <-stopCh:
			s.cfg.polling = false
		case <-s.trig:
			tmp, err := s.Temperature()
			if err != nil {
				errCh <- err
				continue
			}
			r := readings{
				id:          s.ID(),
				temperature: strconv.FormatFloat(float64(tmp), 'f', -1, 32),
				timestamp:   time.Now(),
			}
			data <- r
		}
	}
	// For sure there won't be more data
	close(data)
	if s.cfg.pollType == async {
		s.cfg.ready.Close()
		close(s.trig)
	}

	// sensor created channel (and is the sender side), so should close
	close(errCh)
	// Notify user that we are done
	finCh <- struct{}{}
	close(finCh)
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

func (r readings) Get() (id string, temperature string, timestamp time.Time) {
	return r.id, r.temperature, r.timestamp
}

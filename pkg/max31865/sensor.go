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
	Temperature() (string, error)
	Poll(data chan Readings, pollTime time.Duration) (err error)
}

var _ Sensor = &Max{}

type Max struct {
	ReadWriteCloser
	cfg  *config
	r    *rtd
	trig chan struct{}
	fin  chan struct{}
	stop chan struct{}
	data chan Readings
}

func (m *Max) Poll(data chan Readings, pollTime time.Duration) (err error) {
	if m.cfg.polling.Load() {
		return ErrAlreadyPolling
	}

	m.cfg.polling.Store(true)
	if pollTime == -1 {
		err = m.prepareAsyncPoll()
	} else {
		err = m.prepareSyncPoll(pollTime)
	}

	if err != nil {
		m.cfg.polling.Store(false)
		return err
	}

	m.fin = make(chan struct{})
	m.stop = make(chan struct{})
	m.data = data
	go m.poll()

	return nil
}

func (m *Max) prepareSyncPoll(pollTime time.Duration) error {
	m.trig = make(chan struct{})
	go func(s *Max, pollTime time.Duration) {
		for s.cfg.polling.Load() {
			<-time.After(pollTime)
			if s.cfg.polling.Load() {
				s.trig <- struct{}{}
			}
		}
		close(s.trig)
	}(m, pollTime)

	return nil
}

func (m *Max) prepareAsyncPoll() error {
	if m.cfg.ready == nil {
		return ErrNoReadyInterface
	}
	m.trig = make(chan struct{}, 1)
	return m.cfg.ready.Open(callback, m)
}

func (m *Max) poll() {
	for m.cfg.polling.Load() {
		select {
		case <-m.stop:
			m.cfg.polling.Store(false)
		case <-m.trig:
			tmp, err := m.Temperature()
			m.data <- Readings{
				ID:          m.ID(),
				Temperature: tmp,
				Stamp:       time.Now(),
				Error:       err,
			}
		}
	}
	// For sure there won't be more data
	close(m.data)
	if m.cfg.pollType == async {
		m.cfg.ready.Close()
		close(m.trig)
	}

	// Notify user that we are done
	m.fin <- struct{}{}
	close(m.fin)
}

func callback(args any) error {
	s, ok := args.(*Max)
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

func (m *Max) Temperature() (string, error) {
	r, err := m.read(regConf, regFault+1)
	if err != nil {
		//	can't do much about it
		return "", err
	}
	err = m.r.update(r[regRtdMsb], r[regRtdLsb])
	if err != nil {
		// Not handling error here, should have happened on previous call
		_ = m.clearFaults()
		// make error more specific
		err = fmt.Errorf("%w: errorReg: %v, posibble causes: %v", err, r[regFault], errorCauses(r[regFault], m.cfg.wiring))
		return "", err
	}
	rtd := m.r.rtd()
	tmp := rtdToTemperature(rtd, m.cfg.refRes, m.cfg.rNominal)

	return strconv.FormatFloat(float64(tmp), 'f', -1, 32), nil
}

func (m *Max) Close() error {
	if m.cfg.polling.Load() {
		m.stop <- struct{}{}
		// Close stop channel, not needed anymore
		close(m.stop)
		// Unblock poll
		for range m.data {
		}
		// Wait until finish
		for range m.fin {
		}
	}

	return m.ReadWriteCloser.Close()
}

func (m *Max) ID() string {
	return string(m.cfg.id)
}

func newSensor(options ...Option) (*Max, error) {
	max := &Max{
		ReadWriteCloser: nil,
		r:               newRtd(),
		cfg:             newConfig(),
	}
	for _, opt := range options {
		if err := opt(max); err != nil {
			return nil, err
		}
	}
	// verify after parsing opts
	if err := max.verify(); err != nil {
		return nil, err
	}

	// Do initial regConfig
	if err := max.config(); err != nil {
		return nil, err
	}

	return max, nil
}

func (m *Max) clearFaults() error {
	return m.write(regConf, []byte{m.cfg.clearFaults()})
}

func (m *Max) config() error {
	err := m.write(regConf, []byte{m.cfg.reg()})
	return err
}

func (m *Max) read(addr byte, len int) ([]byte, error) {
	// We need to create slice with 1 byte more
	w := make([]byte, len+1)
	w[0] = addr
	r, err := m.ReadWrite(w)
	if err != nil {
		return nil, err
	}
	// First byte is useless
	return r[1:], nil
}

func (m *Max) write(addr byte, w []byte) error {
	buf := []byte{addr | 0x80}
	buf = append(buf, w...)
	_, err := m.ReadWrite(buf)
	return err
}

func (m *Max) verify() error {
	// Check if interface exists
	if m.ReadWriteCloser == nil {
		return ErrNoReadWriteCloser
	}
	// Check interface itself
	const size = regFault + 2
	buf := make([]byte, size)
	buf[0] = regConf
	r, err := m.ReadWrite(buf)
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

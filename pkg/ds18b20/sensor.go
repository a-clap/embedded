package ds18b20

import (
	"io"
	"strings"
	"time"
)

type Sensor interface {
	ID() string
	Temperature() (string, error)
	Poll(readings chan<- Readings, stopCh <-chan struct{}, pollTime time.Duration) (finCh <-chan struct{}, errCh <-chan error, err error)
}

type Readings interface {
	Get() (id string, temperature string, timestamp time.Time)
}

type readen struct {
	id, temperature string
	timestamp       time.Time
}

type opener interface {
	Open(name string) (File, error)
}

type sensor struct {
	opener
	id      string
	path    string
	polling bool
}

func newSensor(o opener, id, basePath string) (*sensor, error) {
	s := &sensor{
		opener:  o,
		id:      id,
		path:    basePath + "/" + id + "/temperature",
		polling: false,
	}
	if _, err := s.Temperature(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *sensor) Poll(readings chan<- Readings, stopCh <-chan struct{}, pollTime time.Duration) (finCh <-chan struct{}, errCh <-chan error, err error) {
	if s.polling {
		return nil, nil, ErrAlreadyPolling
	}

	s.polling = true
	finChan := make(chan struct{})
	errChan := make(chan error)

	go s.poll(readings, stopCh, pollTime, finChan, errChan)

	return finChan, errChan, nil
}

func (s *sensor) poll(readings chan<- Readings, stopCh <-chan struct{}, pollTime time.Duration, finCh chan struct{}, errCh chan error) {
	for s.polling {
		select {
		case <-stopCh:
			s.polling = false
		case <-time.After(pollTime):
			tmp, err := s.Temperature()
			if err != nil {
				errCh <- err
				continue
			}
			r := readen{
				id:          s.ID(),
				temperature: tmp,
				timestamp:   time.Now(),
			}
			readings <- r
		}
	}
	close(readings)
	// For sure there won't be more data
	// sensor created channel (and is the sender side), so should close
	close(errCh)
	// Notify user that we are done
	finCh <- struct{}{}
	close(finCh)
}

func (s *sensor) Temperature() (string, error) {
	f, err := s.Open(s.path)
	if err != nil {
		return "", err
	}
	// sensor temperature file is just few bytes, io.ReadAll is fine for that purpose
	buf, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}
	conv := strings.TrimRight(string(buf), "\r\n")
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
	return conv, nil
}

func (s *sensor) ID() string {
	return s.id
}

func (r readen) Get() (id string, temperature string, timestamp time.Time) {
	return r.id, r.temperature, r.timestamp
}

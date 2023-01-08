package ds18b20

import (
	"io"
	"strings"
	"time"
)

type Sensor interface {
	io.Closer
	ID() string
	Temperature() (string, error)
	Poll(readings chan Readings, pollTime time.Duration) (err error)
}

var _ Sensor = &sensor{}

type opener interface {
	Open(name string) (File, error)
}

type sensor struct {
	opener
	id      string
	path    string
	polling bool
	fin     chan struct{}
	stop    chan struct{}
	data    chan Readings
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

func (s *sensor) Poll(data chan Readings, pollTime time.Duration) (err error) {
	if s.polling {
		return ErrAlreadyPolling
	}

	s.polling = true
	s.fin = make(chan struct{})
	s.stop = make(chan struct{})
	s.data = data
	go s.poll(pollTime)

	return nil
}

func (s *sensor) Close() error {
	if s.polling {
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

func (s *sensor) poll(pollTime time.Duration) {

	for s.polling {
		select {
		case <-s.stop:
			s.polling = false
		case <-time.After(pollTime):
			tmp, err := s.Temperature()
			s.data <- Readings{
				ID:          s.id,
				Temperature: tmp,
				Stamp:       time.Now(),
				Error:       err,
			}
		}
	}
	close(s.data)
	// For sure there won't be more data
	// sensor created channel (and is the sender side), so should close
	close(s.fin)
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

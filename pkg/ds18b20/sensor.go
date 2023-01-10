package ds18b20

import (
	"io"
	"path"
	"strconv"
	"strings"
	"time"
)

type opener interface {
	Open(name string) (File, error)
}

type Sensor struct {
	opener
	id              string
	temperaturePath string
	resolutionPath  string
	polling         bool
	fin             chan struct{}
	stop            chan struct{}
	data            chan Readings
}

type Resolution int

const (
	Resolution_9_BIT  = 9
	Resolution_10_BIT = 10
	Resolution_11_BIT = 11
	Resolution_12_BIT = 12
)

func newSensor(o opener, id, basePath string) (*Sensor, error) {
	s := &Sensor{
		opener:          o,
		id:              id,
		temperaturePath: path.Join(basePath, id, "temperature"),
		resolutionPath:  path.Join(basePath, id, "resolution"),
		polling:         false,
	}
	if _, err := s.Temperature(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Sensor) Poll(data chan Readings, pollTime time.Duration) (err error) {
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

func (s *Sensor) Resolution() (r Resolution, err error) {
	r = Resolution_12_BIT

	res, err := s.Open(s.resolutionPath)
	if err != nil {
		return
	}
	defer res.Close()
	buf, err := io.ReadAll(res)
	if err != nil {
		return
	}
	maybeR, err := strconv.ParseInt(string(buf), 10, 32)
	if err != nil {
		return
	}

	r = Resolution(maybeR)
	return
}

func (s *Sensor) SetResolution(res Resolution) error {
	resFile, err := s.Open(s.resolutionPath)
	if err != nil {
		return err
	}
	defer resFile.Close()

	buf := strconv.FormatInt(int64(res), 10)
	if _, err := resFile.Write([]byte(buf)); err != nil {
		return err
	}
	return nil
}

func (s *Sensor) Close() error {
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

func (s *Sensor) poll(pollTime time.Duration) {

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
	// Sensor created channel (and is the sender side), so should close
	close(s.fin)
}

func (s *Sensor) Temperature() (string, error) {
	f, err := s.Open(s.temperaturePath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	// Sensor temperature file is just few bytes, io.ReadAll is fine for that purpose
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

func (s *Sensor) ID() string {
	return s.id
}

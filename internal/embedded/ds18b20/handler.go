/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package ds18b20

import (
	"github.com/a-clap/iot/internal/embedded/models"
	"io"
	"path"
	"strconv"
	"strings"
	"time"
)

type Handler struct {
	FileOpener
	id                              string
	temperaturePath, resolutionPath string
	polling                         bool
	fin, stop                       chan struct{}
	data                            chan Readings
}

func newSensor(o FileOpener, id, basePath string) (*Handler, error) {
	s := &Handler{
		FileOpener:      o,
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

func (s *Handler) Poll(data chan Readings, pollTime time.Duration) (err error) {
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

func (s *Handler) Resolution() (r models.DSResolution, err error) {
	r = models.Resolution12BIT

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

	r = models.DSResolution(maybeR)
	return
}

func (s *Handler) SetResolution(res models.DSResolution) error {
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

func (s *Handler) Close() error {
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

func (s *Handler) poll(pollTime time.Duration) {

	for s.polling {
		select {
		case <-s.stop:
			s.polling = false
		case <-time.After(pollTime):
			tmp, err := s.Temperature()
			s.data <- &readings{
				id:          s.id,
				temperature: tmp,
				stamp:       time.Now(),
				err:         err,
			}
		}
	}
	close(s.data)
	// For sure there won't be more data
	// Handler created channel (and is the sender side), so should close
	close(s.fin)
}

func (s *Handler) Temperature() (string, error) {
	f, err := s.Open(s.temperaturePath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	// Handler temperature file is just few bytes, io.ReadAll is fine for that purpose
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

func (s *Handler) ID() string {
	return s.id
}

type readings struct {
	id, temperature string
	stamp           time.Time
	err             error
}

func (r *readings) ID() string {
	return r.id
}

func (r *readings) Temperature() string {
	return r.temperature
}

func (r *readings) Stamp() time.Time {
	return r.stamp
}

func (r *readings) Error() error {
	return r.err
}

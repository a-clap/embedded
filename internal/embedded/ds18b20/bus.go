/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package ds18b20

import (
	"errors"
	"fmt"
	"github.com/a-clap/iot/internal/embedded/logger"
	"io"
	"time"
)

var Log = logger.Log

var (
	ErrInterface      = errors.New("interface")
	ErrAlreadyPolling = errors.New("sensor is already polling")
	ErrNoSuchID       = errors.New("there is no sensor with provided ID")
)

type File interface {
	io.ReadWriteCloser
}

type FileOpener interface {
	Open(name string) (File, error)
}

type Onewire interface {
	Path() string
	ReadDir(dirname string) ([]string, error)
	FileOpener
}

type Readings interface {
	ID() string
	Temperature() string
	Stamp() time.Time
	Error() error
}
type Bus struct {
	ids []string
	o   Onewire
}

func NewBus(options ...BusOption) (*Bus, error) {
	b := &Bus{}
	for _, opt := range options {
		if err := opt(b); err != nil {
			return nil, err
		}
	}

	if b.o == nil {
		return nil, errors.New("lack of Onewire interface")
	}

	return b, nil
}

func (b *Bus) IDs() ([]string, error) {
	err := b.updateIDs()
	return b.ids, err
}

func (b *Bus) NewSensor(id string) (*Handler, error) {
	ids, err := b.IDs()
	if err != nil {
		return nil, err
	}

	found := false
	for _, elem := range ids {
		if elem == id {
			found = true
			break
		}
	}

	if !found {
		return nil, ErrNoSuchID
	}

	// delegate creation of Handler to newSensor
	s, err := newSensor(b.o, id, b.o.Path())
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (b *Bus) updateIDs() error {
	fileNames, err := b.o.ReadDir(b.o.Path())
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInterface, err)
	}
	for _, name := range fileNames {
		if len(name) > 0 {
			// Onewire id starts with digit
			if name[0] >= '0' && name[0] <= '9' {
				b.ids = append(b.ids, name)
			}
		}
	}
	return nil
}

func (b *Bus) Discover() ([]*Handler, error) {
	ids, err := b.IDs()
	if err != nil {
		return nil, err
	}
	s := make([]*Handler, 0, len(ids))
	for _, id := range ids {
		ds, err := b.NewSensor(id)
		if err == nil {
			s = append(s, ds)
		}
	}
	return s, nil
}
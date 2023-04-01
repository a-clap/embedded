/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package ds18b20

import (
	"errors"
	"fmt"
	"io"
)

var (
	ErrNoSuchID    = errors.New("there is no sensor with provided ID")
	ErrNoInterface = errors.New("no interface")
)

type File interface {
	io.ReadWriteCloser
}

// FileOpener is the simplest interface to Open File for read/write
type FileOpener interface {
	Open(name string) (File, error)
}

// Onewire represents Linux onewire driver
type Onewire interface {
	Path() string
	ReadDir(dirname string) ([]string, error)
	FileOpener
}

type Bus struct {
	ids []string
	o   Onewire
}

// NewBus creates Bus with BusOption
// Interface must be presented
func NewBus(options ...BusOption) (*Bus, error) {
	b := &Bus{}
	for _, opt := range options {
		opt(b)
	}

	// Can't do anything without interface
	if b.o == nil {
		return nil, fmt.Errorf("NewBus: %w", ErrNoInterface)
	}

	return b, nil
}

// IDs return slice of DS18B20 ID found on provided Path
func (b *Bus) IDs() ([]string, error) {
	err := b.updateIDs()
	if err != nil {
		return nil, fmt.Errorf("IDs: %w", err)
	}
	return b.ids, nil
}

// NewSensor creates DS18B20 Sensor based on ID
func (b *Bus) NewSensor(id string) (*Sensor, error) {
	// delegate creation of Sensor to NewSensor
	s, err := NewSensor(b.o, id, b.o.Path())
	if err != nil {
		return nil, fmt.Errorf("NewSensor: %w", err)
	}
	return s, nil
}

// Discover create slice of Sensors found on Path
func (b *Bus) Discover() (s []*Sensor, errs []error) {
	ids, err := b.IDs()
	if err != nil {
		return nil, []error{fmt.Errorf("Discover: %w", err)}
	}
	s = make([]*Sensor, 0, len(ids))
	for _, id := range ids {
		ds, err := b.NewSensor(id)
		if err != nil {
			errs = append(errs, fmt.Errorf("Discover.NewSensor: %w", err))
			continue
		}
		s = append(s, ds)
	}
	return s, errs
}

func (b *Bus) updateIDs() error {
	b.ids = nil
	fileNames, err := b.o.ReadDir(b.o.Path())
	if err != nil {
		return fmt.Errorf("ReadDir: {Path: %v}: %w", b.o.Path(), err)
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

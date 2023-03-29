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

// Those are Err possible in Error.Err
var (
	ErrNoSuchID    = errors.New("there is no sensor with provided ID")
	ErrNoInterface = errors.New("no interface")
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

type Bus struct {
	ids []string
	o   Onewire
}

func NewBus(options ...BusOption) (*Bus, error) {
	b := &Bus{}
	for _, opt := range options {
		opt(b)
	}

	if b.o == nil {
		return nil, fmt.Errorf("NewBus: %w", ErrNoInterface)
	}

	return b, nil
}

func (b *Bus) IDs() ([]string, error) {
	err := b.updateIDs()
	if err != nil {
		return nil, fmt.Errorf("IDs {Bus: %s}: %w", b.o.Path(), err)
	}
	return b.ids, nil
}

func (b *Bus) NewSensor(id string) (*Sensor, error) {
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
		return nil, fmt.Errorf("NewSensor {Bus %v, ID: %v}: %w", b.o.Path(), id, ErrNoSuchID)
	}

	// delegate creation of Sensor to NewSensor
	s, err := NewSensor(b.o, id, b.o.Path())
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (b *Bus) Discover() ([]*Sensor, error) {
	ids, err := b.IDs()
	if err != nil {
		return nil, err
	}
	s := make([]*Sensor, 0, len(ids))
	for _, id := range ids {
		ds, err := b.NewSensor(id)
		if err == nil {
			s = append(s, ds)
		}
	}
	return s, nil
}

func (b *Bus) updateIDs() error {
	fileNames, err := b.o.ReadDir(b.o.Path())
	if err != nil {
		return err
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

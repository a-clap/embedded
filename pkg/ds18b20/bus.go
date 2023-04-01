/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package ds18b20

import (
	"errors"
	"fmt"
	"log"
	"path"
	"strings"
)

var (
	ErrNoSuchID    = errors.New("there is no sensor with provided ID")
	ErrNoInterface = errors.New("no interface")
)

type FileReaderWriter interface {
	WriteFile(name string, data []byte) error
	ReadFile(name string) ([]byte, error)
}

// Onewire represents Linux onewire driver
type Onewire interface {
	Path() string
	FileReaderWriter
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
	slavesPath := path.Join(b.o.Path(), "w1_master_slaves")
	ids, err := b.o.ReadFile(slavesPath)
	if err != nil {
		return fmt.Errorf("ReadFile: {Path: %v}: %w", slavesPath, err)
	}
	log.Println(ids)
	b.ids = nil
	if len(ids) == 0 {
		return nil
	}
	b.ids = strings.Split(string(ids), "\n")
	if b.ids[len(b.ids)-1] == "" {
		b.ids = b.ids[:len(b.ids)-1]
	}
	return nil
}

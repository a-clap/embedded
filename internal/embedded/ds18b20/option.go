/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package ds18b20

type BusOption func(bus *Bus) error

func WithInterface(o Onewire) BusOption {
	return func(bus *Bus) error {
		bus.o = o
		return nil
	}
}

func WithOnewire() BusOption {
	return func(bus *Bus) error {
		return WithOnewireOnPath("/sys/bus/w1/devices/w1_bus_master1")(bus)
	}
}

func WithOnewireOnPath(path string) BusOption {
	return func(bus *Bus) error {
		bus.o = &onewire{path: path}
		return nil
	}
}

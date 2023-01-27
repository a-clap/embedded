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

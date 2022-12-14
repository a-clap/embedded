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
		bus.o = &onewire{}
		return nil
	}
}

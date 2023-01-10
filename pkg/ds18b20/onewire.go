package ds18b20

import (
	"os"
)

type onewire struct {
}

func (h *onewire) Path() string {
	return "/sys/bus/w1/devices/w1_bus_master1"
}

func (h *onewire) ReadDir(dirname string) ([]string, error) {
	entries, err := os.ReadDir(dirname)
	if err != nil {
		return nil, err
	}

	dirs := make([]string, len(entries))
	for i, e := range entries {
		dirs[i] = e.Name()
	}
	return dirs, nil
}

func (h *onewire) Open(name string) (File, error) {
	return os.Open(name)
}

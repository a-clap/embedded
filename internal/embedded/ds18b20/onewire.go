package ds18b20

import (
	"os"
)

type onewire struct {
	path string
}

func (o *onewire) Path() string {
	return o.path
}

func (o *onewire) ReadDir(dirname string) ([]string, error) {
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

func (o *onewire) Open(name string) (File, error) {
	return os.Open(name)
}

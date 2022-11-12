package ds18b20

import (
	"io/fs"
	"os"
)

type onewire struct {
}

func (h *onewire) Path() string {
	return "/sys/bus/w1/ids"
}

func (h *onewire) ReadDir(dirname string) ([]fs.DirEntry, error) {
	return os.ReadDir(dirname)
}

func (h *onewire) Open(name string) (File, error) {
	return os.Open(name)
}

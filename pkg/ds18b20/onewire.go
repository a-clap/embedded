package ds18b20

import (
	"io/fs"
	"io/ioutil"
	"os"
)

type onewire struct {
}

func (h *onewire) Path() string {
	return "/sys/bus/w1/ids"
}

func (h *onewire) ReadDir(dirname string) ([]fs.FileInfo, error) {
	return ioutil.ReadDir(dirname)
}

func (h *onewire) Open(name string) (File, error) {
	return os.Open(name)
}

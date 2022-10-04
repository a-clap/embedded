package ds

import (
	"fmt"
	"io"
	"io/fs"
)

var (
	ErrInterface = fmt.Errorf("interface")
)

type File interface {
	io.Reader
}

type Handler interface {
	Path() string
	ReadDir(dirname string) ([]fs.FileInfo, error)
	Open(name string) (File, error)
}

type DS struct {
	h Handler
}

func New(h Handler) *DS {
	return &DS{h: h}
}

func (d *DS) Devices() (devices []string, err error) {
	files, err := d.h.ReadDir(d.h.Path())
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInterface, err)
	}
	for _, maybeOnewire := range files {
		if name := maybeOnewire.Name(); len(name) > 0 {
			// Onewire devices names starts with digit
			if name[0] >= '0' && name[0] < '9' {
				devices = append(devices, maybeOnewire.Name())
			}
		}
	}
	return
}

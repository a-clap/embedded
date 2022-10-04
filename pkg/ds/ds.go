package ds

import (
	"fmt"
	"github.com/a-clap/logger"
	"io/fs"
)

var Log logger.Logger = logger.NewNop()

var (
	ErrInterface = fmt.Errorf("interface")
)

type Handler interface {
	Path() string
	ReadDir(dirname string) ([]fs.FileInfo, error)
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

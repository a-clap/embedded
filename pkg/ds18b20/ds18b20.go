package ds18b20

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
)

var (
	ErrInterface      = errors.New("interface")
	ErrAlreadyPolling = errors.New("sensor is already polling")
)

type File interface {
	io.Reader
}

type Onewire interface {
	Path() string
	ReadDir(dirname string) ([]fs.DirEntry, error)
	Open(name string) (File, error)
}

type Handler struct {
	ids []string
	o   Onewire
}

func New(o Onewire) *Handler {
	return &Handler{
		o:   o,
		ids: nil,
	}
}

func NewDefault() *Handler {
	return &Handler{
		o:   &onewire{},
		ids: nil,
	}
}

func (h *Handler) IDs() ([]string, error) {
	err := h.updateIDs()
	return h.ids, err
}

func (h *Handler) NewSensor(id string) (Sensor, error) {
	// delegate creation of sensor to newSensor
	s, err := newSensor(h.o, id, h.o.Path())
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (h *Handler) updateIDs() error {
	files, err := h.o.ReadDir(h.o.Path())
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInterface, err)
	}
	for _, maybeOnewire := range files {
		if name := maybeOnewire.Name(); len(name) > 0 {
			// Onewire ID starts with digit
			if name[0] >= '0' && name[0] <= '9' {
				h.ids = append(h.ids, name)
			}
		}
	}
	return nil
}

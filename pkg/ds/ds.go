package ds

import (
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"strings"
)

var (
	ErrInterface = fmt.Errorf("interface")
)

type File interface {
	io.Reader
}

type Onewire interface {
	Path() string
	ReadDir(dirname string) ([]fs.FileInfo, error)
	Open(name string) (File, error)
}

type Sensor interface {
	ID() string
	Temperature() (string, error)
}

type opener interface {
	Open(name string) (File, error)
}

type ds struct {
	o    opener
	id   string
	path string
}

type Handler struct {
	devices []string
	o       Onewire
}

func New(o Onewire) *Handler {
	return &Handler{
		o:       o,
		devices: nil,
	}
}

func (h *Handler) Devices() ([]string, error) {
	err := h.updateDevices()
	return h.devices, err
}

func (h *Handler) updateDevices() error {
	files, err := h.o.ReadDir(h.o.Path())
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInterface, err)
	}
	for _, maybeOnewire := range files {
		if name := maybeOnewire.Name(); len(name) > 0 {
			// Onewire devices names starts with digit
			if name[0] >= '0' && name[0] <= '9' {
				h.devices = append(h.devices, name)
			}
		}
	}
	return nil
}

func (h *Handler) NewSensor(id string) (Sensor, error) {
	// We could search array of devices...
	// or just assume that user know, what is doing
	s := &ds{
		o:    h.o,
		id:   id,
		path: h.o.Path() + "/" + id + "/temperature",
	}
	// But let's check whether it is possible to read temperature from DS18B20
	if _, err := s.Temperature(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *ds) Temperature() (string, error) {
	f, err := s.o.Open(s.path)
	if err != nil {
		return "", err
	}
	// ds temperature file is just few bytes, ioutil.ReadAll is fine for that purpose
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}
	conv := strings.TrimRight(string(buf), "\r\n")
	length := len(conv)
	if length > 3 {
		conv = conv[:length-3] + "." + conv[length-3:]
	} else {
		leading := "0."
		for length < 3 {
			leading += "0"
			length++
		}
		conv = leading + conv
	}
	return conv, nil
}

func (s *ds) ID() string {
	return s.id
}

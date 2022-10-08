package ds18b20

import (
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"
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

type Readings interface {
	Get() (id string, temperature string, timestamp time.Time)
}

type data struct {
	id, temperature string
	timestamp       time.Time
}

type opener interface {
	Open(name string) (File, error)
}

type handler struct {
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
func NewDefault() *Handler {

	return &Handler{
		o:       &handler{},
		devices: nil,
	}
}

func (h *Handler) Devices() ([]string, error) {
	err := h.updateDevices()
	return h.devices, err
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

func (h *Handler) Poll(ids []string, readings chan<- Readings, exitCh <-chan struct{}, interval time.Duration) (finChan <-chan struct{}, errCh <-chan error, errors []error) {

	sensors := make([]Sensor, 0, len(ids))
	for _, id := range ids {
		if s, err := h.NewSensor(id); err == nil {
			sensors = append(sensors, s)
		} else {
			errors = append(errors, err)
		}
	}
	// No sensor available
	if len(sensors) == 0 {
		return nil, nil, errors
	}

	fin := make(chan struct{})
	errChan := make(chan error)
	go h.poll(ids, readings, exitCh, fin, errChan, interval)
	return fin, errChan, errors
}

func (h *Handler) poll(ids []string, readings chan<- Readings, exit <-chan struct{}, finChan chan<- struct{}, errCh chan error, interval time.Duration) {
	w := sync.WaitGroup{}
	w.Add(1)

	// Let's make it at least number of sensors
	dataCh := make(chan Readings, len(ids))
	stopCh := make(chan struct{})
	for _, id := range ids {
		if s, err := h.NewSensor(id); err == nil {
			go pollSingle(s, dataCh, stopCh, errCh, interval, &w)
		} else {
			fmt.Println("failed", err)
		}
		// TODO: notify about error
	}

	go func() {
		defer w.Done()
		for {
			select {
			case <-exit:
				close(stopCh)
				return
			case sens := <-dataCh:
				readings <- sens
			}
		}
	}()

	// Waiting for all goroutines
	w.Wait()
	// At this moment, everything should be done
	close(errCh)
	// Notify user that we are done
	finChan <- struct{}{}
	// We are responsible for closing channel, but still - don't have to do that, it will be garbage collected
	close(finChan)
}

func pollSingle(s Sensor, sensorCh chan<- Readings, stopCh <-chan struct{}, errCh chan error, pollTime time.Duration, w *sync.WaitGroup) {
	w.Add(1)
	defer func() {
		w.Done()
	}()

	r := data{
		id: s.ID(),
	}

	for {
		select {
		case <-stopCh:
			return
		default:
		}

		select {
		case <-stopCh:
			return
		case <-time.After(pollTime):
			tmp, err := s.Temperature()
			if err != nil {
				errCh <- err
				return
			}
			r.timestamp = time.Now()
			r.temperature = tmp
			sensorCh <- r
		}
	}
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

func (d data) Get() (id string, temperature string, timestamp time.Time) {
	return d.id, d.temperature, d.timestamp
}

func (h *handler) Path() string {
	return "/sys/bus/w1/devices"
}

func (h *handler) ReadDir(dirname string) ([]fs.FileInfo, error) {
	return ioutil.ReadDir(dirname)
}

func (h *handler) Open(name string) (File, error) {
	return os.Open(name)
}

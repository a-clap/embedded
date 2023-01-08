package ds18b20

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

type BusHandler struct {
	discover Discover
	sensors  []Sensor
	polling  atomic.Bool

	channels   []chan Readings
	close      chan struct{}
	waitClosed chan struct{}
}

type Discover interface {
	Discover() ([]Sensor, error)
}

type BusHandlerOption func(*BusHandler) error

func NewBusHandler(option ...BusHandlerOption) (*BusHandler, error) {
	b := &BusHandler{}
	for _, opt := range option {
		if err := opt(b); err != nil {
			return nil, err
		}
	}
	if err := b.init(); err != nil {
		return nil, err
	}

	return b, nil
}

func (b *BusHandler) Handle(readings chan<- Readings, interval time.Duration) (int, error) {
	if b.polling.Load() {
		return 0, errors.New("already polling")
	}

	return b.prepare(readings, interval)
}

func (b *BusHandler) Close() error {
	if !b.polling.Load() {
		return errors.New("not polling")
	}
	close(b.close)
	for range b.waitClosed {
	}
	return nil
}

func (b *BusHandler) prepare(readings chan<- Readings, interval time.Duration) (int, error) {
	var err error
	b.sensors, err = b.discover.Discover()
	if err != nil {
		return 0, err
	}

	b.channels = make([]chan Readings, len(b.sensors))
	good := len(b.sensors)
	for i, ds := range b.sensors {
		if err := ds.Poll(b.channels[i], interval); err != nil {
			good -= 1
		}
	}
	if good > 0 {
		b.close = make(chan struct{})
		b.waitClosed = make(chan struct{})
		go b.handle(readings)
	}
	return good, nil
}

func (b *BusHandler) handle(readings chan<- Readings) {
	wait := sync.WaitGroup{}

	// Pass data to readings from each channel
	for _, ch := range b.channels {
		go func(ch chan Readings, group *sync.WaitGroup) {
			wait.Add(1)
			for data := range ch {
				readings <- data
			}
			wait.Done()
		}(ch, &wait)
	}

	for b.polling.Load() {
		// wait until closed
		for range b.close {
		}
		// then close every sensor
		for _, s := range b.sensors {
			_ = s.Close()
		}

		b.polling.Store(false)
	}

	wait.Wait()
	close(readings)
	close(b.waitClosed)
}

func (b *BusHandler) init() error {
	if b.discover == nil {
		return errors.New("lack of Discover interface")
	}
	return nil
}

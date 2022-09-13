package ds18b20

import (
	"github.com/a-clap/logger"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var Log logger.Logger = logger.NewNop()

const (
	// W1Path path, where we can find onewire devices
	W1Path = "/sys/bus/w1/devices"
)

// Sensor return information about current polled sensor
type Sensor struct {
	ID, Temperature string
}

// ParseTemperature returns s.Temperature in format X.YYY degrees
func (s Sensor) ParseTemperature() (float32, error) {
	// DS returns temperature with 3 digits after dot
	conv := s.Temperature
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
	f, err := strconv.ParseFloat(conv, 32)
	if err != nil {
		return 0, err
	}
	return float32(f), nil

}

// SensorsIDs returns list of detected sensors
func SensorsIDs() (ids []string, err error) {
	Log.Debug("Reading directory: ", W1Path)
	files, err := ioutil.ReadDir(W1Path)
	if err != nil {
		Log.Error("Error on reading directory: ", err)
		return nil, err
	}

	for _, file := range files {
		if name := file.Name(); len(name) > 0 {
			Log.Debug("Found file: ", name)
			// Sensor ID starts with digit, they maybe also w1_master_slave
			if name[0] >= '0' && name[0] <= '9' {
				ids = append(ids, file.Name())
			}
		}
	}
	return ids, nil
}

// PollAll starts poll on every device found in user space
// for rest parameters check Poll
func PollAll(readings chan<- Sensor, exitCh <-chan struct{}, interval time.Duration) (finChan <-chan struct{}, errCh <-chan error, err error) {
	ids, err := SensorsIDs()
	if err != nil {
		return nil, nil, err
	}
	return Poll(ids, readings, exitCh, interval)
}

// Poll starts polling on devices in ids[]
// readings is user channel, where he will receive Sensor structure
// exitCh provides way of finishing polling
// interval is time between two consecutive reads
// finChan is information channel, user should read than channel after exitCh <-, so he can be sure that job is done
// errCh will send errors from polling
func Poll(ids []string, readings chan<- Sensor, exitCh <-chan struct{}, interval time.Duration) (finChan <-chan struct{}, errCh <-chan error, err error) {
	fin := make(chan struct{})
	errChan := make(chan error)
	go poll(ids, readings, exitCh, fin, errChan, interval)

	return fin, errChan, nil
}

func poll(ids []string, readings chan<- Sensor, exit <-chan struct{}, finChan chan<- struct{}, errCh chan error, interval time.Duration) {
	// If some jerk try to kill us with same IDs...
	ids = removeDuplicates(ids)

	w := sync.WaitGroup{}
	w.Add(1)

	dataCh := make(chan Sensor)
	stopCh := make(chan struct{})
	for _, id := range ids {
		go pollSingle(id, dataCh, stopCh, errCh, interval, &w)
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

func pollSingle(id string, sensorCh chan<- Sensor, stopCh <-chan struct{}, errCh chan error, pollTime time.Duration, w *sync.WaitGroup) {
	w.Add(1)
	defer func() {
		Log.Info("Closing device...")
		w.Done()
	}()

	path := W1Path + "/" + id + "/temperature"
	s := Sensor{
		ID:          id,
		Temperature: "",
	}

	sendErr := func(err error) {
		errCh <- err
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
			f, err := os.Open(path)
			if err != nil {
				Log.Warn("Error on opening file ", path, ", err is ", err)
				sendErr(err)
				// TODO: Should we return here?
				return
			}

			// We don't expect that there will be a lot of data, 10 bytes maybe max, so I feel it is safe to use ReadAll
			buf, err := ioutil.ReadAll(f)
			if err != nil {
				Log.Warn("Error on reading file ", path, ", err is ", err)
				sendErr(err)
				// TODO: Should we return here?
				return
			}
			s.Temperature = strings.TrimRight(string(buf), "\r\n")
			Log.Debug("Got reading: ", s)
			sensorCh <- s

			err = f.Close()
			if err != nil {
				Log.Warn("Error on closing file ", path, ", err is ", err)
				sendErr(err)
				return
			}
		}
	}
}

// removeDuplicates remove duplicate elements from slice
// Generics just 4fun
func removeDuplicates[T comparable](s []T) []T {
	keys := make(map[T]bool)
	var list []T
	for _, elem := range s {
		if _, ok := keys[elem]; !ok {
			list = append(list, elem)
			keys[elem] = true
		}
	}
	return list
}

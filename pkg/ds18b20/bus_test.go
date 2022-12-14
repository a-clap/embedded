package ds18b20_test

import (
	"fmt"
	"github.com/a-clap/iot/internal/models"
	"github.com/a-clap/iot/pkg/ds18b20"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type iError struct {
}

func (i *iError) Open(string) (ds18b20.File, error) {
	panic("shouldn't be used")
}

func (i *iError) ReadDir(string) ([]fs.DirEntry, error) {
	return nil, fmt.Errorf("interfaceError")
}

func (i *iError) Path() string {
	return ""
}

type iAfero struct {
	path string
	a    afero.IOFS
}

func (i *iAfero) ReadDir(dirname string) ([]fs.DirEntry, error) {
	return i.a.ReadDir(dirname)
}

func (i *iAfero) Open(name string) (ds18b20.File, error) {
	if len(name) > 1 && name[0] == '/' {
		name = name[1:]
	}
	return i.a.Open(name)
}

func (i *iAfero) Path() string {
	return i.path
}

var _ ds18b20.Onewire = &iError{}
var _ ds18b20.Onewire = &iAfero{}

func TestHandler_Devices(t *testing.T) {
	var err error
	af := afero.Afero{Fs: afero.NewMemMapFs()}

	nodevice := "/empty"
	require.Nil(t, af.Mkdir(nodevice, 0777))

	justMasterDevice := "/just/master/device"

	require.Nil(t, af.Mkdir(justMasterDevice, 0777))
	_, err = af.Create(justMasterDevice + "/w1_bus_master")
	require.Nil(t, err)

	singleOneWireDevicePath := "/onewire/single_device"
	singleOneWireDevice := "28-05169397aeff"
	require.Nil(t, af.Mkdir(singleOneWireDevicePath, 0777))
	_, err = af.Create(singleOneWireDevicePath + "/" + singleOneWireDevice)
	require.Nil(t, err)

	multipleDevicesPath := "/onewire/multiple_devices"
	multipleDevices := []string{"1234", "182-2313123", "999996696"}
	require.Nil(t, af.Mkdir(multipleDevicesPath, 0777))
	for _, device := range multipleDevices {
		_, err = af.Create(multipleDevicesPath + "/" + device)
	}

	require.Nil(t, err)

	defer func() { _ = af.RemoveAll("") }()

	tests := []struct {
		name    string
		handler ds18b20.Onewire
		want    []string
		wantErr bool
		err     error
	}{
		{
			name:    "handle interface error",
			handler: &iError{},
			want:    nil,
			wantErr: true,
			err:     ds18b20.ErrInterface,
		},
		{
			name: "no onewire device",
			handler: &iAfero{
				path: nodevice,
				a:    afero.NewIOFS(af),
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "just master device",
			handler: &iAfero{
				path: justMasterDevice,
				a:    afero.NewIOFS(af),
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "single one wire device",
			handler: &iAfero{
				path: singleOneWireDevicePath,
				a:    afero.NewIOFS(af),
			},
			want:    []string{singleOneWireDevice},
			wantErr: false,
		},
		{
			name: "multiple ids",
			handler: &iAfero{
				path: multipleDevicesPath,
				a:    afero.NewIOFS(af),
			},
			want:    multipleDevices,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, _ := ds18b20.NewBus(ds18b20.WithInterface(tt.handler))
			got, err := h.IDs()

			if tt.wantErr {
				require.NotNil(t, err)
				require.ErrorIs(t, err, tt.err)
				return
			}

			require.Nil(t, err)
			require.EqualValues(t, tt.want, got)
		})
	}
}

func TestHandler_NewSensor(t *testing.T) {
	af := afero.Afero{Fs: afero.NewMemMapFs()}
	defer func() { _ = af.RemoveAll("") }()

	sensorDoesntexist := "not_exist"

	sensorIDWithoutTemperaturePath := "/exist"
	sensorIDWithoutTemperature := "81-12131"
	p := filepath.Join(sensorIDWithoutTemperaturePath, sensorIDWithoutTemperature)
	require.Nil(t, af.Mkdir(p, 0777))

	sensorGoodID := "28-abcdefg"
	sensorGoodPath := "good"
	p = filepath.Join(sensorGoodPath, sensorGoodID)
	require.Nil(t, af.Mkdir(p, 0777))

	p = filepath.Join(p, "temperature")
	f, err := af.Create(p)
	require.Nil(t, err)
	sensorGoodTemperature := "98121"

	_, err = f.Write([]byte(sensorGoodTemperature))
	require.Nil(t, err)
	_ = f.Close()

	tests := []struct {
		name    string
		o       ds18b20.Onewire
		argsId  string
		wantErr bool
		errType error
	}{
		{
			name: "sensor doesn't exist",
			o: &iAfero{
				path: sensorDoesntexist,
				a:    afero.NewIOFS(af),
			},
			argsId:  "blabla",
			wantErr: true,
			errType: os.ErrNotExist,
		},
		{
			name: "temperature file doesn't exist",
			o: &iAfero{
				path: sensorIDWithoutTemperaturePath,
				a:    afero.NewIOFS(af),
			},
			argsId:  sensorIDWithoutTemperature,
			wantErr: true,
			errType: os.ErrNotExist,
		},
		{
			name: "working sensor",
			o: &iAfero{
				path: sensorGoodPath,
				a:    afero.NewIOFS(af),
			},
			argsId:  sensorGoodID,
			wantErr: false,
			errType: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, _ := ds18b20.NewBus(ds18b20.WithInterface(tt.o))
			s, err := h.NewSensor(tt.argsId)

			if tt.wantErr {
				require.NotNil(t, err)
				require.ErrorIs(t, err, tt.errType)
				return
			}

			require.Nil(t, err)
			require.EqualValues(t, tt.argsId, s.ID())
		})
	}
}

func TestHandler_SensorTemperature(t *testing.T) {
	af := afero.Afero{Fs: afero.NewMemMapFs()}
	defer func() { _ = af.RemoveAll("") }()
	// Prepare files
	id := "28-12313asb"
	p := filepath.Join("wire", id)
	require.Nil(t, af.Mkdir(p, 0777))
	filePath := filepath.Join(p, "temperature")
	_, err := af.Create(filePath)
	require.Nil(t, err)
	err = af.Chmod(filePath, 0777)
	require.Nil(t, err)
	// Prepare interface
	o := &iAfero{
		path: "wire",
		a:    afero.NewIOFS(af),
	}
	// Get sensor tested
	h, _ := ds18b20.NewBus(ds18b20.WithInterface(o))
	s, err := h.NewSensor(id)
	require.Nil(t, err)
	require.Equal(t, id, s.ID())

	tests := []struct {
		write    string
		expected string
	}{
		{
			write:    "1",
			expected: "0.001",
		},
		{
			write:    "988654\r\n",
			expected: "988.654",
		},
		{
			write:    "12355\r\n",
			expected: "12.355",
		},
		{
			write:    "1230\r",
			expected: "1.230",
		},
		{
			write:    "456\n",
			expected: "0.456",
		},
		{
			write:    "38\n",
			expected: "0.038",
		},
	}

	t.Run("proper conversions", func(t *testing.T) {
		for _, test := range tests {
			f, err := af.OpenFile(filePath, os.O_WRONLY, 0777)
			require.Nil(t, err)
			n, err := f.Write([]byte(test.write))
			require.Equal(t, len(test.write), n)
			require.Nil(t, err)
			_ = f.Close()

			r, err := s.Temperature()
			require.Nil(t, err)

			require.EqualValues(t, test.expected, r)
		}
	})
}

func TestSensor_PollTwice(t *testing.T) {
	af := afero.Afero{Fs: afero.NewMemMapFs()}
	defer func() { _ = af.RemoveAll("") }()
	// Prepare sensor
	expectedID := "281ab"
	tmp := "12345"
	require.Nil(t, af.Mkdir(expectedID, 0777))
	f, err := af.Create(expectedID + "/temperature")
	require.Nil(t, err)

	_, _ = f.Write([]byte(tmp))
	_ = f.Close()

	o := &iAfero{
		path: "",
		a:    afero.NewIOFS(af),
	}

	readings := make(chan models.SensorReadings)
	interval := 5 * time.Millisecond
	h, _ := ds18b20.NewBus(ds18b20.WithInterface(o))
	s, _ := h.NewSensor(expectedID)

	errs := s.Poll(readings, interval)
	require.Nil(t, errs)

	err = s.Poll(readings, interval)
	require.ErrorIs(t, err, ds18b20.ErrAlreadyPolling)

	wait := make(chan struct{})

	go func() {
		require.Nil(t, s.Close())
		wait <- struct{}{}
	}()

	select {
	case <-wait:
	case <-time.After(2 * interval):
		require.Fail(t, "should be done after this time")
	}
	close(wait)

}

func TestHandler_Poll_IntervalsTemperatureUpdate(t *testing.T) {
	af := afero.Afero{Fs: afero.NewMemMapFs()}
	defer func() { _ = af.RemoveAll("") }()
	// Prepare sensor
	expectedID := "281ab"
	expectedTemp := "12.345"
	tmp := "12345"
	require.Nil(t, af.Mkdir(expectedID, 0777))
	f, err := af.Create(expectedID + "/temperature")
	require.Nil(t, err)

	_, _ = f.Write([]byte(tmp))
	_ = f.Close()

	o := &iAfero{
		path: "",
		a:    afero.NewIOFS(af),
	}

	readings := make(chan models.SensorReadings)
	interval := 5 * time.Millisecond
	h, _ := ds18b20.NewBus(ds18b20.WithInterface(o))
	s, _ := h.NewSensor(expectedID)

	errs := s.Poll(readings, interval)
	require.Nil(t, errs)

	for i := 0; i < 10; i++ {
		now := time.Now()
		select {
		case r := <-readings:
			require.EqualValues(t, expectedID, r.ID)
			require.EqualValues(t, expectedTemp, r.Temperature)
			require.Nil(t, r.Error)
			diff := r.Stamp.Sub(now)
			require.Less(t, interval, diff)
			require.InDelta(t, interval.Milliseconds(), diff.Milliseconds(), float64(interval.Milliseconds())/10)
		case <-time.After(2 * interval):
			require.Fail(t, "failed, waiting for readings too long")
		}
	}
	wait := make(chan struct{})

	go func() {
		require.Nil(t, s.Close())
		wait <- struct{}{}
	}()

	select {
	case <-wait:
	case <-time.After(2 * interval):
		require.Fail(t, "should be done after this time")
	}

}

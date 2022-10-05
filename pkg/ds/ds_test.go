package ds_test

import (
	"fmt"
	"github.com/a-clap/beaglebone/pkg/ds"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

type iError struct {
}

func (i *iError) Open(name string) (ds.File, error) {
	panic("shouldn't be used")
}

func (i *iError) ReadDir(dirname string) ([]fs.FileInfo, error) {
	return nil, fmt.Errorf("interfaceError")
}

func (i *iError) Path() string {
	return ""
}

type iAfero struct {
	path string
	a    *afero.Afero
}

func (i *iAfero) ReadDir(dirname string) ([]fs.FileInfo, error) {
	return i.a.ReadDir(dirname)
}

func (i *iAfero) Open(name string) (ds.File, error) {
	return i.a.Open(name)
}

func (i *iAfero) Path() string {
	return i.path
}

var _ ds.Onewire = &iError{}
var _ ds.Onewire = &iAfero{}

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

	defer af.RemoveAll("/*")

	tests := []struct {
		name    string
		handler ds.Onewire
		want    []string
		wantErr bool
		err     error
	}{
		{
			name:    "handle interface error",
			handler: &iError{},
			want:    nil,
			wantErr: true,
			err:     ds.ErrInterface,
		},
		{
			name: "no onewire device",
			handler: &iAfero{
				path: nodevice,
				a:    &af,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "just master device",
			handler: &iAfero{
				path: justMasterDevice,
				a:    &af,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "single one wire device",
			handler: &iAfero{
				path: singleOneWireDevicePath,
				a:    &af,
			},
			want:    []string{singleOneWireDevice},
			wantErr: false,
		},
		{
			name: "multiple devices",
			handler: &iAfero{
				path: multipleDevicesPath,
				a:    &af,
			},
			want:    multipleDevices,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := ds.New(tt.handler)
			got, err := h.Devices()

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
	defer af.RemoveAll("/")

	sensorDoesntexist := "/not_exist"

	sensorIDWithoutTemperaturePath := "/exist"
	sensorIDWithoutTemperature := "81-12131"
	p := filepath.Join(sensorIDWithoutTemperaturePath, sensorIDWithoutTemperature)
	require.Nil(t, af.Mkdir(p, 0777))

	sensorGoodID := "28-abcdefg"
	sensorGoodPath := "/good"
	p = filepath.Join(sensorGoodPath, sensorGoodID)
	require.Nil(t, af.Mkdir(p, 0777))

	p = filepath.Join(p, "temperature")
	f, err := af.Create(p)
	require.Nil(t, err)
	sensorGoodTemperature := "98121"

	_, err = f.Write([]byte(sensorGoodTemperature))
	require.Nil(t, err)
	f.Close()

	tests := []struct {
		name    string
		o       ds.Onewire
		argsId  string
		wantErr bool
		errType error
	}{
		{
			name: "sensor doesn't exist",
			o: &iAfero{
				path: sensorDoesntexist,
				a:    &af,
			},
			argsId:  "blabla",
			wantErr: true,
			errType: os.ErrNotExist,
		},
		{
			name: "temperature file doesn't exist",
			o: &iAfero{
				path: sensorIDWithoutTemperaturePath,
				a:    &af,
			},
			argsId:  sensorIDWithoutTemperature,
			wantErr: true,
			errType: os.ErrNotExist,
		},
		{
			name: "working sensor",
			o: &iAfero{
				path: sensorGoodPath,
				a:    &af,
			},
			argsId:  sensorGoodID,
			wantErr: false,
			errType: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := ds.New(tt.o)
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
	defer af.RemoveAll("/")
	// Prepare files
	id := "28-12313asb"
	p := filepath.Join("/wire", id)
	require.Nil(t, af.Mkdir(p, 0777))
	filePath := filepath.Join(p, "temperature")
	_, err := af.Create(filePath)
	require.Nil(t, err)
	// Prepare interface
	o := &iAfero{
		path: "/wire",
		a:    &af,
	}
	// Get sensor tested
	s, err := ds.New(o).NewSensor(id)
	require.Nil(t, err)
	require.Equal(t, id, s.ID())

	tests := []struct {
		write    string
		expected string
	}{
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
		{
			write:    "1",
			expected: "0.001",
		},
	}

	t.Run("proper conversions", func(t *testing.T) {
		for _, test := range tests {
			err := af.WriteFile(filePath, []byte(test.write), 0644)
			require.Nil(t, err)

			r, err := s.Temperature()
			require.Nil(t, err)

			require.EqualValues(t, test.expected, r)
		}
	})
}

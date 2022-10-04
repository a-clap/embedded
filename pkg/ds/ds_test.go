package ds_test

import (
	"fmt"
	"github.com/a-clap/beaglebone/pkg/ds"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"io/fs"
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

var _ ds.Handler = &iError{}
var _ ds.Handler = &iAfero{}

func TestHandler_Devices(t *testing.T) {
	var err error
	af := afero.Afero{Fs: afero.NewMemMapFs()}

	justMasterDevice := "/just/master/device"

	require.Nil(t, af.Mkdir(justMasterDevice, 0777))
	_, err = af.Create(justMasterDevice + "/w1_bus_master")
	require.Nil(t, err)

	singleOneWireDevicePath := "/onewire/single_device"
	singleOneWireDevice := "28-05169397aeff"
	require.Nil(t, af.Mkdir(singleOneWireDevicePath, 0777))
	_, err = af.Create(singleOneWireDevicePath + "/" + singleOneWireDevice)
	require.Nil(t, err)

	defer af.RemoveAll("/*")

	tests := []struct {
		name    string
		handler ds.Handler
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := ds.New(tt.handler)
			got, err := h.Devices()

			if tt.wantErr {
				require.NotNil(t, err)
				require.ErrorContains(t, err, tt.err.Error())
				return
			}

			require.Nil(t, err)
			require.EqualValues(t, tt.want, got)
		})
	}
}

package max31865_test

import (
	"fmt"
	"github.com/a-clap/beaglebone/pkg/max31865"
	"github.com/stretchr/testify/require"
	"testing"
)

type transfer struct {
	val byte
	err error
}

func (t transfer) Close() error {
	return nil
}

func (t transfer) ReadWrite(write []byte) ([]byte, error) {
	if t.err != nil {
		return nil, t.err
	}
	size := len(write)
	r := make([]byte, size)
	for i := 0; i < size; i++ {
		r[i] = t.val
	}
	return r, nil
}

var _ max31865.Transfer = &transfer{}

func TestNew(t *testing.T) {

	tests := []struct {
		name     string
		c        max31865.Config
		transfer max31865.Transfer
		wantErr  bool
		errType  error
	}{
		{
			name: "all good",
			c: max31865.Config{
				Wiring:   max31865.FourWire,
				RefRes:   430.0,
				RNominal: 100.0,
			},
			transfer: transfer{val: 1, err: nil},
			wantErr:  false,
			errType:  nil,
		},
		{
			name: "interface error",
			c: max31865.Config{
				Wiring:   max31865.FourWire,
				RefRes:   430.0,
				RNominal: 100.0,
			},
			transfer: transfer{val: 1, err: fmt.Errorf("interface error")},
			wantErr:  true,
			errType:  max31865.ErrReadWrite,
		},
		{
			name: "only zeroes",
			c: max31865.Config{
				Wiring:   max31865.FourWire,
				RefRes:   430.0,
				RNominal: 100.0,
			},
			transfer: transfer{val: 0, err: nil},
			wantErr:  true,
			errType:  max31865.ErrReadZeroes,
		},
		{
			name: "only ff",
			c: max31865.Config{
				Wiring:   max31865.FourWire,
				RefRes:   430.0,
				RNominal: 100.0,
			},
			transfer: transfer{val: 0xff, err: nil},
			wantErr:  true,
			errType:  max31865.ErrReadFF,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := max31865.New(tt.transfer, tt.c)

			if tt.wantErr {
				require.NotNil(t, err)
				require.ErrorIs(t, err, tt.errType)
				return
			}
			require.Nil(t, err)

		})
	}
}

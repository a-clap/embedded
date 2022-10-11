package max_test

import (
	"fmt"
	"github.com/a-clap/beaglebone/pkg/max"
	"github.com/stretchr/testify/require"
	"testing"
)

type transfer struct {
	val byte
	err error
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

var _ max.Transfer = &transfer{}

func TestNew(t *testing.T) {

	tests := []struct {
		name     string
		c        max.Config
		transfer max.Transfer
		wantErr  bool
		errType  error
	}{
		{
			name: "all good",
			c: max.Config{
				Wiring:   max.FourWire,
				RefRes:   430.0,
				RNominal: 100.0,
			},
			transfer: transfer{val: 1, err: nil},
			wantErr:  false,
			errType:  nil,
		},
		{
			name: "interface error",
			c: max.Config{
				Wiring:   max.FourWire,
				RefRes:   430.0,
				RNominal: 100.0,
			},
			transfer: transfer{val: 1, err: fmt.Errorf("interface error")},
			wantErr:  true,
			errType:  max.ErrReadWrite,
		},
		{
			name: "only zeroes",
			c: max.Config{
				Wiring:   max.FourWire,
				RefRes:   430.0,
				RNominal: 100.0,
			},
			transfer: transfer{val: 0, err: nil},
			wantErr:  true,
			errType:  max.ErrReadZeroes,
		},
		{
			name: "only ff",
			c: max.Config{
				Wiring:   max.FourWire,
				RefRes:   430.0,
				RNominal: 100.0,
			},
			transfer: transfer{val: 0xff, err: nil},
			wantErr:  true,
			errType:  max.ErrReadFF,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := max.New(tt.transfer, tt.c)

			if tt.wantErr {
				require.NotNil(t, err)
				require.ErrorIs(t, err, tt.errType)
				return
			}
			require.Nil(t, err)

		})
	}
}

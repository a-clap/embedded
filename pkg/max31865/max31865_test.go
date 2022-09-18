package max31865

import (
	"errors"
	"fmt"
	"github.com/a-clap/logger"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"testing"
)

func init() {
	Log = logger.NewDefaultZap(zapcore.DebugLevel)
}

type data struct {
	w   []byte
	r   []byte
	err error
}

type fakeTransfer struct {
	i    int
	data []data
}

func (f *fakeTransfer) TxRx(w []byte) (r []byte, err error) {
	defer func() { f.i++ }()
	f.data[f.i].w = w
	return f.data[f.i].r, f.data[f.i].err
}

var _ Transfer = &fakeTransfer{}

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		t       fakeTransfer
		want    [][]byte
		config  Config
		wantErr bool
		err     error
	}{
		{
			name: "handle interface error",
			t: fakeTransfer{
				i: 0,
				data: []data{
					{
						w:   nil,
						r:   nil,
						err: fmt.Errorf("hello error"),
					},
				},
			},
			config:  Config{},
			wantErr: true,
			err:     ErrInterface,
		},
		{
			name: "new writes updates config, then write 1",
			t: fakeTransfer{
				i: 0,
				data: []data{
					{
						w:   nil,
						r:   make([]byte, 9),
						err: nil,
					},
					{},
				},
			},
			want: [][]byte{
				{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
				{0x80, 0b10010000},
			},
			config: Config{
				Bias:       true,
				AutoMode:   false,
				Wire3:      true,
				Filter50Hz: false,
				RefRes:     100,
				RNominal:   430,
				reg:        0,
			},
			wantErr: false,
			err:     nil,
		},
		{
			name: "new writes updates config, then write 2",
			t: fakeTransfer{
				i: 0,
				data: []data{
					{
						w:   nil,
						r:   make([]byte, 9),
						err: nil,
					},
					{},
				},
			},
			want: [][]byte{
				{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
				{0x80, 0b11010001},
			},
			config: Config{
				Bias:       true,
				AutoMode:   true,
				Wire3:      true,
				Filter50Hz: true,
				RefRes:     100,
				RNominal:   430,
				reg:        0,
			},
			wantErr: false,
			err:     nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(&tt.t, tt.config)
			if tt.wantErr {
				require.NotNil(t, err)
				require.True(t, errors.Is(err, tt.err))
				return
			}

			require.Nil(t, err)
			require.Equal(t, 2, tt.t.i)
			for i, elem := range tt.t.data {
				require.Equal(t, tt.want[i], elem.w)
			}
		})
	}
}

func TestDevice_Temperature(t *testing.T) {
	tests := []struct {
		name    string
		t       fakeTransfer
		want    [][]byte
		wantErr bool
		err     error
	}{
		{
			name: "clear faults, if received rtd fault",
			t: fakeTransfer{
				i: 0,
				data: []data{
					{
						w:   nil,
						r:   make([]byte, 9),
						err: nil,
					},
					{
						w:   nil,
						r:   make([]byte, 9),
						err: nil,
					},
					{
						w: nil,
						r: []byte{
							0x1, 0x0, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1,
						},
						err: nil,
					},
					{
						w:   nil,
						r:   make([]byte, 9),
						err: nil,
					},
				},
			},
			want: [][]byte{
				nil,         // don't check it
				nil,         // don't check it
				nil,         // don't check it
				{0x80, 0x2}, // here should be clear fault with last read config
			},
			wantErr: false,
			err:     nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := NewDefault(&tt.t)
			require.Nil(t, err)

			got, err := d.Temperature()
			if tt.wantErr {
				require.NotNil(t, err)
				require.True(t, errors.Is(err, ErrFault))
			}
			for i, elem := range tt.t.data {
				if tt.want[i] != nil {
					require.Equal(t, tt.want[i], elem.w)
				}
			}
			require.NotNil(t, got)

		})
	}
}

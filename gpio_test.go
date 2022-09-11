package beaglebone_test

import (
	"github.com/a-clap/beaglebone"
	"github.com/a-clap/logger"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"testing"
)

func init() {
	beaglebone.Log = logger.NewDefaultZap(zapcore.DebugLevel)
}

func TestBeagleboneGpio_DigitalWriteRead(t *testing.T) {
	type args struct {
		pinName string
		value   byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		err     error
	}{
		{
			name: "write read",
			args: args{
				pinName: "1",
				value:   0,
			},
			wantErr: false,
			err:     nil,
		},
		{
			name: "can't assign pin, which is not a number",
			args: args{
				pinName: "abc",
				value:   0,
			},
			wantErr: true,
			err:     beaglebone.ErrCantParsePin,
		},
		{
			name: "assign a pin over 31 - as it exceeds gpiochip0",
			args: args{
				pinName: "32",
				value:   1,
			},
			wantErr: false,
			err:     nil,
		},
		{
			name: "assign a pin over 63 - as it exceeds gpiochip1",
			args: args{
				pinName: "64",
				value:   1,
			},
			wantErr: false,
			err:     nil,
		},
		{
			name: "assign a pin over 95 - as it exceeds gpiochip2",
			args: args{
				pinName: "96",
				value:   1,
			},
			wantErr: false,
			err:     nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := beaglebone.NewAdaptor()
			err := b.DigitalWrite(tt.args.pinName, tt.args.value)
			if tt.wantErr {
				require.NotNil(t, err)
				require.Equal(t, tt.err, err)
				return
			}
			require.Nil(t, err)

			val, err := b.DigitalRead(tt.args.pinName)

			require.Nil(t, err)
			require.Equal(t, tt.args.value, byte(val))
		})
	}

	t.Run("multiple write reads", func(t *testing.T) {
		b := beaglebone.NewAdaptor()
		// usrled0
		pin := "22"
		for i := 0; i < 5; i++ {
			val := byte(i & 0x1)
			err := b.DigitalWrite(pin, val)
			require.Nil(t, err)

			readVal, err := b.DigitalRead(pin)
			require.Nil(t, err)
			require.Equal(t, val, byte(readVal))
		}
	})
}

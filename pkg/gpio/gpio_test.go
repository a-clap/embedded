package gpio_test

import (
	"github.com/a-clap/beaglebone/pkg/gpio"
	"github.com/a-clap/logger"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"testing"
)

func init() {
	gpio.Log = logger.NewDefaultZap(zapcore.DebugLevel)
}

func TestInput(t *testing.T) {
	type args struct {
		pin int
	}
	type wants struct {
		err      bool
		offset   int
		chipName string
	}
	tests := []struct {
		name  string
		args  args
		wants wants
	}{
		{
			name: "get random pin for chip0",
			args: args{
				pin: 21,
			},
			wants: wants{
				err:      false,
				offset:   21,
				chipName: "gpiochip0",
			},
		},
		{
			name: "get random pin for chip1",
			args: args{
				pin: 51,
			},
			wants: wants{
				err:      false,
				offset:   51 % 32,
				chipName: "gpiochip1",
			},
		},
		{
			name: "get random pin for chip2",
			args: args{
				pin: 84,
			},
			wants: wants{
				err:      false,
				offset:   84 % 32,
				chipName: "gpiochip2",
			},
		},
		{
			name: "get random pin for chip3",
			args: args{
				pin: 111,
			},
			wants: wants{
				err:      false,
				offset:   111 % 32,
				chipName: "gpiochip3",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := gpio.Input(tt.args.pin)
			if tt.wants.err {
				require.NotNil(t, err)
				return
			}
			defer got.Close()

			require.Nil(t, err)

			require.Equal(t, tt.wants.offset, got.Offset())
			require.Equal(t, tt.wants.chipName, got.Chip())
		})
	}
}

func TestOutput(t *testing.T) {
	type args struct {
		pin       int
		initValue bool
	}
	type wants struct {
		err      bool
		offset   int
		chipName string
	}
	tests := []struct {
		name  string
		args  args
		wants wants
	}{
		{
			name: "get random pin for chip0",
			args: args{
				pin:       21,
				initValue: false,
			},
			wants: wants{
				err:      false,
				offset:   21,
				chipName: "gpiochip0",
			},
		},
		{
			name: "get random pin for chip1",
			args: args{
				pin:       51,
				initValue: true,
			},
			wants: wants{
				err:      false,
				offset:   51 % 32,
				chipName: "gpiochip1",
			},
		},
		{
			name: "get random pin for chip2",
			args: args{
				pin:       84,
				initValue: false,
			},
			wants: wants{
				err:      false,
				offset:   84 % 32,
				chipName: "gpiochip2",
			},
		},
		{
			name: "get random pin for chip3",
			args: args{
				pin:       111,
				initValue: true,
			},
			wants: wants{
				err:      false,
				offset:   111 % 32,
				chipName: "gpiochip3",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := gpio.Output(tt.args.pin, tt.args.initValue)
			if tt.wants.err {
				require.NotNil(t, err)
				return
			}
			defer got.Close()

			require.Nil(t, err)
			require.Equal(t, tt.wants.offset, got.Offset())
			require.Equal(t, tt.wants.chipName, got.Chip())
			val, err := got.Get()
			require.Nil(t, err)
			require.Equal(t, tt.args.initValue, val)
		})
	}
}

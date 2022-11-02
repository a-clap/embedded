package ws2812_test

import (
	"fmt"
	"github.com/a-clap/iot/pkg/ws2812"
	"github.com/stretchr/testify/require"
	"testing"
)

type WS2821Writer struct {
	bytes []byte
	err   error
}

func (w *WS2821Writer) Write(bytes []byte) error {
	w.bytes = bytes
	return w.err
}

func TestWS2812_SetColor(t *testing.T) {
	tests := []struct {
		name    string
		size    uint
		idx     uint
		wantErr bool
		err     error
	}{
		{
			name:    "all good",
			size:    1,
			idx:     0,
			wantErr: false,
			err:     nil,
		},
		{
			name:    "idx over size",
			size:    1,
			idx:     1,
			wantErr: true,
			err:     ws2812.ErrLedNotExist,
		},
		{
			name:    "size equals to 0, useless struct",
			size:    0,
			idx:     0,
			wantErr: true,
			err:     ws2812.ErrLedNotExist,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := ws2812.New(tt.size, &WS2821Writer{})
			// not really care about rgb
			err := w.SetColor(tt.idx, 0, 0, 0)
			if tt.wantErr {
				require.NotNil(t, err)
				require.ErrorIs(t, err, tt.err)
				return
			}
			require.Nil(t, err)
		})
	}
}

func TestWS2812_Refresh(t *testing.T) {

	t.Run("interface error", func(t *testing.T) {
		w := ws2812.New(1, &WS2821Writer{err: fmt.Errorf("interface error")})
		_ = w.SetColor(0, 0, 0, 0)
		err := w.Refresh()
		require.NotNil(t, err)
		require.ErrorIs(t, err, ws2812.ErrInterface)
	})

	t.Run("internal buffer len corresponds to size", func(t *testing.T) {
		sizes := []uint{0, 5, 13, 122, 13}
		for i, size := range sizes {
			writer := WS2821Writer{}
			w := ws2812.New(size, &writer)
			err := w.Refresh()
			require.Nil(t, err)
			// 3 bytes for reset signal
			// for each led there are 24 bit
			expectedSize := 3 + 24*size
			require.Len(t, writer.bytes, int(expectedSize), "failed on pos %v, size %v", i, size)
		}
	})

	t.Run("buffer starts with {0,0,0}", func(t *testing.T) {
		writer := WS2821Writer{}
		w := ws2812.New(1, &writer)
		_ = w.SetColor(0, 255, 255, 255)
		err := w.Refresh()
		require.Nil(t, err)
		for i := 0; i < 3; i++ {
			require.Equal(t, byte(0), writer.bytes[i])
		}
	})

	t.Run("SetColor affects specified led", func(t *testing.T) {
		writer := WS2821Writer{}
		w := ws2812.New(5, &writer)
		require.Nil(t, w.Refresh())

		// get buffer written, no matter what is inside
		b := make([]byte, 0, len(writer.bytes))
		_ = copy(b, writer.bytes)
		for i := uint(0); i < 5; i++ {

			require.Nil(t, w.SetColor(i, 255, 255, 255))
			require.Nil(t, w.Refresh())

			touchedLedStart := int(i*24 + 3)
			touchedLedEnd := int((i+1)*24 + 3)

			newBuf := make([]byte, 0, len(writer.bytes))
			_ = copy(newBuf, writer.bytes)
			for pos, elem := range newBuf {
				switch {
				// first three elements are 0
				case pos < 3:
					require.Equal(t, byte(0), elem)
				// Touched led should be different
				case pos >= touchedLedStart && pos < touchedLedEnd:
					require.NotEqual(t, b[pos], elem)
				case pos >= len(b):
					require.FailNow(t, "something gone bad")
				//	Rest stays equal
				default:
					require.Equal(t, b[pos], elem)
				}
			}
			// update b with new values
			b = newBuf
		}
	})

	t.Run("buffer is handled properly", func(t *testing.T) {
		const (
			zero byte = 0b11000000
			one  byte = 0b11111000
		)
		writer := WS2821Writer{}
		w := ws2812.New(1, &writer)
		r, g, b := uint8(255), uint8(255), uint8(0)

		require.Nil(t, w.SetColor(0, r, g, b))
		require.Nil(t, w.Refresh())

		parseColor := func(pos int, color uint8) uint8 {
			if (color & (1 << pos)) == 0 {
				return zero
			}
			return one
		}

		greenBytes := 1*8 + 3 - 1
		redBytes := 2*8 + 3 - 1
		blueBytes := 3*8 + 3 - 1

		for pos, elem := range writer.bytes {
			switch {
			// first three elements are 0
			case pos < 3:
				require.Equal(t, byte(0), elem)
			case pos <= greenBytes:
				exp := parseColor(greenBytes-pos, g)
				require.Equal(t, exp, elem)
			case pos <= redBytes:
				exp := parseColor(redBytes-pos, r)
				require.Equal(t, exp, elem)
			case pos <= blueBytes:
				exp := parseColor(blueBytes-pos, b)
				require.Equal(t, exp, elem)
			default:
				require.FailNow(t, "shouldn't be here")
			}
			// update b with new values
		}
	})
}

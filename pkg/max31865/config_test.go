package max31865

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestConfig_reg(t *testing.T) {
	tests := []struct {
		name        string
		wiring      Wiring
		reg         uint8
		clearFaults uint8
		faultDetect uint8
	}{
		{
			name:        "two wire",
			wiring:      TwoWire,
			reg:         0b11000001,
			clearFaults: 0b11000001 | (1 << 1),
			faultDetect: 0b10000101,
		},
		{
			name:        "four wire",
			wiring:      FourWire,
			reg:         0b11000001,
			clearFaults: 0b11000001 | (1 << 1),
			faultDetect: 0b10000101,
		},
		{
			name:        "three wire",
			wiring:      ThreeWire,
			reg:         0b11010001,
			clearFaults: 0b11010001 | (1 << 1),
			faultDetect: 0b10010101,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newConfig(tt.wiring)

			if got := c.reg(); got != tt.reg {
				t.Errorf("reg() = %v, reg %v", got, tt.reg)

			}
			if got := c.clearFaults(); got != tt.clearFaults {
				t.Errorf("clearFaults() = %v, clearFaults %v", got, tt.clearFaults)
			}

			if got := c.faultDetect(); got != tt.faultDetect {
				t.Errorf("FaultDetect() = %v, faultDetect %v", got, tt.faultDetect)
			}
		})
	}
}

func TestConfig_faultDetectFinished(t *testing.T) {

	wireModes := []Wiring{TwoWire, ThreeWire, FourWire}
	regRunning := []uint8{0b11111111, 0b10101010, 0b00000100, 0b00001000, 0b10001000}
	regFinished := []uint8{0b11110011, 0b10100010, 0b00000000, 0b10100011, 0b10000000}

	t.Run("test many options", func(t *testing.T) {
		for i, wiring := range wireModes {
			c := newConfig(wiring)

			for k, running := range regRunning {
				require.Falsef(t, c.faultDetectFinished(running), "should detect running for wiring = %s, and regRunning %v", wireModes[i], regRunning[k])
			}

			for k, finished := range regFinished {
				require.True(t, c.faultDetectFinished(finished), "should detect finish for wiring = \"%s\", and regFinished \"%v\"", wireModes[i], regFinished[k])
			}

		}
	})

}

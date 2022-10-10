package max

import "testing"

func TestConfig_reg(t *testing.T) {
	tests := []struct {
		name   string
		wiring Wiring
		want   uint8
	}{
		{
			name:   "two wire",
			wiring: TwoWire,
			want:   0b11000001,
		},
		{
			name:   "four wire",
			wiring: FourWire,
			want:   0b11000001,
		},
		{
			name:   "three wire",
			wiring: ThreeWire,
			want:   0b11010001,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewConfig(tt.wiring)

			if got := c.reg(); got != tt.want {
				t.Errorf("reg() = %v, want %v", got, tt.want)
			}
		})
	}
}

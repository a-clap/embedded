package numeric

import "strings"

type implFloat struct {
	Value
	clr     bool
	current string
}

func (v *implFloat) Enter() {
	// Check, if user passed dot somewhere
	if strings.Contains(v.current, ".") {
		if v.current[len(v.current)-1] == '.' {
			v.current += "0"
		}
	} else {
		v.current += ".0"
	}
	v.Value.Set(v.current)
}

var _ impl = &implFloat{}

func newFloat(v Value) *implFloat {
	return &implFloat{
		clr:     true,
		Value:   v,
		current: v.Get(),
	}
}

func (v *implFloat) Clear() {
	v.current = "0.0"
}
func (v *implFloat) Backspace() {
	n := len(v.current)
	if n <= 1 {
		v.current = "0"
		return
	}
	v.current = v.current[:n-1]
}

func (v *implFloat) Dot() {
	if !strings.Contains(v.current, ".") {
		v.current += "."
	}
}

func (v *implFloat) Digit(d string) {
	if v.current == "0.0" && v.clr {
		v.current = ""
		v.clr = false
	}
	v.current += d
}

func (v *implFloat) Get() string {
	return v.current
}

func (v *implFloat) Minus() {
	if len(v.current) == 0 {
		return
	}

	if v.current[0] == '-' {
		v.current = v.current[1:]
		return
	}
	v.current = "-" + v.current
}

package max

import "fmt"

var (
	twoWireErrors = [...]string{
		"Overvoltage or undervoltage fault",                  // D2
		"RTDIN- shorted low, RTDIN+ shorted low",             // D3
		"RTDIN- shorted low",                                 // D4
		"Open RTD, RTDIN+ shorted high, RTDIN- shorted high", // D5
		"Shorted RTD element, RTDIN+ shorted low",            // D6
		"Open RTD element",                                   // D7
	}

	threeWireErrors = [...]string{
		"Overvoltage or undervoltage fault",                                               // D2
		"Force+ shorted low, RTDIN+ shorted low and connected to RTD, RTDIN- shorted low", // D3
		"RTDIN- shorted low", // D4
		"Open RTD element, Force+ shorted high and connected to RTD, Force+ unconnected, Force+ shorted high and not connected to RTD, RTDIN- shorted high", // D5
		"RTDIN+ shorted to RTDIN-, RTDIN+ shorted low and not connected to RTD, Force+ shorted low",                                                         // D6
		"Open RTD element, RTDIN+ shorted high and not connected to RTD, Force+ shorted high and connected to RTD",                                          // D7
	}

	fourWireErrors = [...]string{
		"Overvoltage or undervoltage fault", // D2
		"Force+ shorted low, RTDIN+ shorted low and connected to RTD, RTDIN- shorted low and connected to RTD, RTDIN- shorted low and not, Force- shorted low", // D3
		"Force- shorted low and connected to RTD, RTDIN- shorted low and connected to RTD",                                                                     // D4
		"Open RTD element, Force+ shorted high and connected to RTD, Force- unconnected, Force+ unconnected, Force+ shorted high and not connected to RTD, " +
			"Force- shorted high and not connected to RTD, Force- shorted high and connected to RTD, Force- shorted low and not connected to RTD", // D5
		"RTDIN+ shorted to RTDIN-, RTDIN+ shorted low and not connected to RTD, RTDIN- shorted high and not connected to RTD, Force+ shorted low", // D6
		"Open RTD element, RTDIN+ shorted high and not connected to RTD, Force+ shorted high and connected to RTD",                                // D7
	}
)

func getError(errorReg byte, w Wiring) error {
	const offset = 2
	bitPos := offset
	var s []string
	for bitPos < 8 {
		if errorReg&0x1 == 0x1 {
			var info string
			pos := bitPos - offset
			switch w {
			case TwoWire:
				info = twoWireErrors[pos]
			case ThreeWire:
				info = threeWireErrors[pos]
			case FourWire:
				info = fourWireErrors[pos]
			}
			s = append(s, info)
		}
		bitPos++
		errorReg >>= 1
	}

	return fmt.Errorf("%w: errorReg: %v, info: %v", ErrRtd, errorReg, s)
}

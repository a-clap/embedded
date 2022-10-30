package gpio

import (
	"fmt"
)

var (
	ErrNotExist = fmt.Errorf("not exist")
)

var BananaPI = map[string]Pin{
	"PWR_LED":  {"gpiochip1", 10},
	"CON2_P03": {"gpiochip0", 12},
	"CON2_P05": {"gpiochip0", 11},
	"CON2_P07": {"gpiochip0", 6},
	"CON2_P08": {"gpiochip0", 13},
	"CON2_P10": {"gpiochip0", 14},
	"CON2_P11": {"gpiochip0", 1},
	"CON2_P12": {"gpiochip0", 16},
	"CON2_P13": {"gpiochip0", 0},
	"CON2_P15": {"gpiochip0", 3},
	"CON2_P16": {"gpiochip0", 15},
	"CON2_P18": {"gpiochip0", 68},
	"CON2_P19": {"gpiochip0", 64},
	"CON2_P21": {"gpiochip0", 65},
	"CON2_P22": {"gpiochip0", 2},
	"CON2_P23": {"gpiochip0", 66},
	"CON2_P24": {"gpiochip0", 67},
	"CON2_P26": {"gpiochip0", 71},
	"CON2_P27": {"gpiochip0", 19},
	"CON2_P28": {"gpiochip0", 18},
	"CON2_P29": {"gpiochip0", 7},
	"CON2_P31": {"gpiochip0", 8},
	"CON2_P32": {"gpiochip1", 2},
	"CON2_P33": {"gpiochip0", 9},
	"CON2_P35": {"gpiochip0", 10},
	"CON2_P36": {"gpiochip1", 4},
	"CON2_P37": {"gpiochip0", 17},
	"CON2_P38": {"gpiochip0", 21},
	"CON2_P40": {"gpiochip0", 20},
}

func BananaPiPin(name string) (Pin, error) {
	maybePin, ok := BananaPI[name]
	if !ok {
		return Pin{}, ErrNotExist
	}
	return maybePin, nil
}

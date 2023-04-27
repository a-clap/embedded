/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package embedded

import (
	"github.com/a-clap/logging"
)

var (
	logger = logging.GetLogger()
)

type Embedded struct {
	Heaters *HeaterHandler
	DS      *DSHandler
	PT      *PTHandler
	GPIO    *GPIOHandler
}

func New(options ...Option) (*Embedded, error) {
	e := &Embedded{
		Heaters: new(HeaterHandler),
		DS:      new(DSHandler),
		PT:      new(PTHandler),
		GPIO:    new(GPIOHandler),
	}

	for _, opt := range options {
		if err := opt(e); err != nil {
			return nil, err
		}
	}

	e.Heaters.Open()
	e.DS.Open()
	e.PT.Open()
	e.GPIO.Open()

	return e, nil
}

func (e *Embedded) close() {
	e.Heaters.Close()
	e.DS.Close()
	e.PT.Close()
	e.GPIO.Close()
}

func Parse(c Config) ([]Option, []error) {
	var errs []error
	var opts []Option
	{
		heaterOpts, err := parseHeaters(c.Heaters)
		if err != nil {
			logger.Error("parseHeaters failed")
			errs = append(errs, err...)
		}

		if heaterOpts != nil {
			opts = append(opts, heaterOpts)
		}
	}
	{
		dsOpts, err := parseDS18B20(c.DS18B20)
		if err != nil {
			logger.Error("parseDS18B20 failed")
			errs = append(errs, err...)
		}
		if dsOpts != nil {
			opts = append(opts, dsOpts)
		}
	}
	{
		ptOpts, err := parsePT100(c.PT100)
		if err != nil {
			logger.Error("parsePT100 failed")
			errs = append(errs, err...)
		}
		if ptOpts != nil {
			opts = append(opts, ptOpts)
		}
	}
	{
		gpioOpts, err := parseGPIO(c.GPIO)
		if err != nil {
			logger.Error("parseGPIO failed")
			errs = append(errs, err...)
		}
		if gpioOpts != nil {
			opts = append(opts, gpioOpts)
		}
	}

	return opts, errs
}

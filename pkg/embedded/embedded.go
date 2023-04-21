/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package embedded

import (
	"github.com/a-clap/logging"
	"github.com/gin-gonic/gin"
)

var (
	logger = logging.GetLogger()
)

type Embedded struct {
	*gin.Engine
	Heaters *HeaterHandler
	DS      *DSHandler
	PT      *PTHandler
	GPIO    *GPIOHandler
}

func New(options ...Option) (*Embedded, error) {
	h := &Embedded{
		Engine:  gin.Default(),
		Heaters: new(HeaterHandler),
		DS:      new(DSHandler),
		PT:      new(PTHandler),
		GPIO:    new(GPIOHandler),
	}
	
	for _, opt := range options {
		if err := opt(h); err != nil {
			return nil, err
		}
	}
	
	h.Heaters.Open()
	h.DS.Open()
	h.PT.Open()
	h.GPIO.Open()
	
	h.routes()
	return h, nil
}

func (e *Embedded) Close() {
	e.Heaters.Close()
	e.DS.Close()
	e.PT.Close()
	e.GPIO.Close()
}

func NewFromConfig(c Config) (*Embedded, error) {
	var opts []Option
	{
		heaterOpts, err := parseHeaters(c.Heaters)
		if err != nil {
			logger.Error("parseHeaters failed")
		}
		
		if heaterOpts != nil {
			opts = append(opts, heaterOpts)
		}
	}
	{
		dsOpts, err := parseDS18B20(c.DS18B20)
		if err != nil {
			logger.Error("parseDS18B20 failed")
		}
		if dsOpts != nil {
			opts = append(opts, dsOpts)
		}
	}
	{
		ptOpts, err := parsePT100(c.PT100)
		if err != nil {
			logger.Error("parsePT100 failed")
		}
		if ptOpts != nil {
			opts = append(opts, ptOpts)
		}
	}
	{
		gpioOpts, err := parseGPIO(c.GPIO)
		if err != nil {
			logger.Error("parseGPIO failed")
		}
		if gpioOpts != nil {
			opts = append(opts, gpioOpts)
		}
	}
	return New(opts...)
}

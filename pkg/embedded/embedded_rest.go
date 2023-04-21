/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package embedded

import (
	"github.com/gin-gonic/gin"
)

type Rest struct {
	Router *restRouter
	*Embedded
}

func NewRest(options ...Option) (*Rest, error) {
	r := &Rest{
		Router: &restRouter{Engine: gin.Default()},
	}
	e, err := New(options...)
	if err != nil {
		return nil, err
	}
	r.Embedded = e
	r.Router.routes(r.Embedded)
	
	return r, nil
}

func (r *Rest) Run() error {
	return r.Router.Run()
}

func (r *Rest) Close() {
	r.Embedded.close()
}

func NewFromConfig(c Config) (*Rest, error) {
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
	return NewRest(opts...)
}

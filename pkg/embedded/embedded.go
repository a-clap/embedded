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
	url     string
}

func New(options ...Option) (*Embedded, error) {
	e := &Embedded{
		Heaters: new(HeaterHandler),
		DS:      new(DSHandler),
		PT:      new(PTHandler),
		GPIO:    new(GPIOHandler),
		url:     "8080",
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

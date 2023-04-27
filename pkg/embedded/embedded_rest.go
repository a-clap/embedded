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
	url    string
	*Embedded
}

func NewRest(url string, options ...Option) (*Rest, error) {
	r := &Rest{
		Router: &restRouter{Engine: gin.Default()},
		url:    url,
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
	return r.Router.Run(r.url)
}

func (r *Rest) Close() {
	r.Embedded.close()
}

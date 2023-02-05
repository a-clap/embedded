/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package rest_test

import (
	"github.com/gin-gonic/gin"
	"io"
)

func init() {
	gin.SetMode(gin.TestMode)
	gin.DefaultWriter = io.Discard
}

/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package main

import "github.com/a-clap/iot/internal/rest"

func main() {
	w, err := New()
	if err != nil {
		panic(err)
	}

	srv, err := rest.New(rest.WithWifiHandler(w))
	if err != nil {
		panic(err)
	}

	panic(srv.Run())
}

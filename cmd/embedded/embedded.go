/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package main

import (
	"flag"
	"log"
	"strconv"

	"github.com/a-clap/embedded/pkg/embedded"
)

var (
	port = flag.Int("port", 50001, "the server port")
	rest = flag.Bool("rest", false, "use REST API instead of gRPC")
)

type handler interface {
	Run() error
	Close()
}

func main() {
	flag.Parse()

	opts, errs := getOpts()
	if errs != nil {
		log.Println(errs)
	}
	if opts == nil || len(opts) == 0 {
		log.Fatalln("Can't run without any option")
	}
	addr := "localhost:" + strconv.FormatInt(int64(*port), 10)

	var err error
	var handler handler
	if *rest {
		log.Println("Running embedded as REST server on ", addr)
		handler, err = embedded.NewRest(addr, opts...)
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		log.Println("Running embedded as RPC server on ", addr)
		handler, err = embedded.NewRPC(addr, opts...)
		if err != nil {
			log.Fatalln(err)
		}

	}
	err = handler.Run()
	log.Println(err)
}

/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package main

import (
	"flag"
	"log"
	"strconv"
	"time"

	"github.com/a-clap/embedded/pkg/embedded"
)

var (
	port = flag.Int("port", 50001, "the server port")
	rpc  = flag.Bool("rest", false, "use REST API instead of gRPC")
)

type handler interface {
	Run() error
	Close()
}

func main() {
	flag.Parse()

	opts, errs := getOpts()
	if errs != nil {
		log.Fatalln(errs)
	}
	addr := "localhost:" + strconv.FormatInt(int64(*port), 10)

	var err error
	var handler handler
	if *rpc {
		log.Println("Running embedded as RPC server on ", addr)
		handler, err = embedded.NewRPC("localhost:50001", opts...)
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		log.Println("Running embedded as REST server on ", addr)
		handler, err = embedded.NewRest(addr, opts...)
		if err != nil {
			log.Fatalln(err)
		}
	}
	go func() {
		<-time.After(time.Second)
		restClients()
	}()
	err = handler.Run()
	log.Println(err)
}

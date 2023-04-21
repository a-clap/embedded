/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package main

import (
	"log"
	
	"github.com/a-clap/embedded/pkg/embedded"
	"github.com/spf13/viper"
)

func main() {
	handler := getEmbeddedFromConfig()
	err := handler.Run()
	log.Println(err)
}

func getEmbeddedFromConfig() *embedded.Embedded {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
	
	cfg := embedded.Config{}
	if err := viper.Unmarshal(&cfg); err != nil {
		panic(err)
	}
	
	e, err := embedded.NewFromConfig(cfg)
	if err != nil {
		panic(err)
	}
	return e
}

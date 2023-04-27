package main

import (
	"github.com/a-clap/embedded/pkg/embedded"
	"github.com/spf13/viper"
)

func getOpts() ([]embedded.Option, []error) {
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
	
	return embedded.Parse(cfg)
}

package main

import (
	"github.com/a-clap/embedded/pkg/embedded"
	"github.com/a-clap/logging"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func setupLogging() {
	logger := logging.GetLogger()

	cfg := zap.NewProductionEncoderConfig()
	rot := logging.RotateOptions{
		MaxSize:    128,
		MaxAge:     31,
		MaxBackups: 2,
		LocalTime:  false,
		Compress:   true,
	}
	logger.AddRotateFileHandler(cfg, logging.DebugLevel, "/var/log/embedded/embedded.log", rot)

	dsLogger := logging.GetLogger("ds18b20")
	dsLogger.AddRotateFileHandler(cfg, logging.DebugLevel, "/var/log/embedded/ds18b20.log", rot)

	maxLogger := logging.GetLogger("max31865")
	maxLogger.AddRotateFileHandler(cfg, logging.DebugLevel, "/var/log/embedded/max31865.log", rot)

	gpioLogger := logging.GetLogger("gpio")
	gpioLogger.AddRotateFileHandler(cfg, logging.DebugLevel, "/var/log/embedded/gpio.log", rot)

	heaterLogger := logging.GetLogger("heater")
	heaterLogger.AddRotateFileHandler(cfg, logging.DebugLevel, "/var/log/embedded/heater.log", rot)

}

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

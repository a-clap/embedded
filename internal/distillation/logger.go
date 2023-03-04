/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package distillation

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Errorf(format string, args ...any)
	Error(args ...any)

	Fatalf(format string, args ...any)
	Fatal(args ...any)

	Infof(format string, args ...any)
	Info(args ...any)

	Warnf(format string, args ...any)
	Warn(args ...any)

	Debugf(format string, args ...any)
	Debug(args ...any)
}

func NewDefaultLogger(level zapcore.Level) *zap.SugaredLogger {
	cfg := zap.NewDevelopmentConfig()
	cfg.Level = zap.NewAtomicLevelAt(level)
	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000")
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	log, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	return log.Sugar()
}

func NewNopLogger() *zap.SugaredLogger {
	return zap.NewNop().Sugar()
}

var log Logger = NewDefaultLogger(zapcore.DebugLevel)

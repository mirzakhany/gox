package common

import (
	"log"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewServiceLogger(level, serviceName, serviceVersion string, opts ...zap.Option) *zap.Logger {
	return NewLogger(level, opts...).With(zap.String("service", serviceName), zap.String("version", serviceVersion))
}

func NewLogger(level string, opts ...zap.Option) *zap.Logger {
	var logLevel zapcore.Level
	if err := logLevel.Set(level); err != nil {
		log.Fatal(err)
	}

	atom := zap.NewAtomicLevel()
	atom.SetLevel(logLevel)

	ops := []zap.Option{zap.ErrorOutput(zapcore.Lock(os.Stderr)), zap.AddCaller()}
	ops = append(ops, opts...)

	logger := zap.New(zapcore.NewSamplerWithOptions(zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.Lock(os.Stdout),
		atom,
	), time.Second, 100, 10),
		ops...,
	)

	return logger
}

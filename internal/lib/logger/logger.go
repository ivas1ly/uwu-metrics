package logger

import (
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(level string) *zap.Logger {
	var logLevel zapcore.Level

	switch strings.ToLower(level) {
	case "debug":
		logLevel = zap.DebugLevel
	case "info":
		logLevel = zap.InfoLevel
	case "warn":
		logLevel = zap.WarnLevel
	case "error":
		logLevel = zap.ErrorLevel
	case "fatal":
		logLevel = zap.FatalLevel
	default:
		logLevel = zap.InfoLevel
	}

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.RFC3339NanoTimeEncoder

	zapConfig := zap.Config{
		Level:             zap.NewAtomicLevelAt(logLevel),
		Development:       false,
		DisableCaller:     false,
		DisableStacktrace: false,
		Sampling:          nil,
		Encoding:          "json",
		EncoderConfig:     encoderCfg,
		OutputPaths:       []string{"stdout"},
		ErrorOutputPaths:  []string{"stdout"},
		InitialFields:     nil,
	}

	logger := zap.Must(zapConfig.Build())
	defer logger.Sync()

	zap.ReplaceGlobals(logger)

	return logger
}

package logger

import (
	"bytes"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type MemorySink struct {
	*bytes.Buffer
}

func (s *MemorySink) Close() error { return nil }
func (s *MemorySink) Sync() error  { return nil }

func TestLogger(t *testing.T) {
	sink := &MemorySink{new(bytes.Buffer)}
	err := zap.RegisterSink("memory", func(*url.URL) (zap.Sink, error) {
		return sink, nil
	})
	assert.NoError(t, err)

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.RFC3339NanoTimeEncoder

	loggerCfg := NewDefaultLoggerConfig()
	loggerCfg.ErrorOutputPaths = []string{"memory://"}
	loggerCfg.OutputPaths = []string{"memory://"}

	testDebugMsg := "test debug message"
	testInfoMsg := "test info message"
	testWarnMsg := "test warn message"
	testErrorMsg := "test error message"
	testFatalMsg := "test fatal message"

	t.Run("test debug level", func(t *testing.T) {
		log := New("debug", loggerCfg)

		log.Info(testInfoMsg)
		outputInfo := sink.String()
		assert.Contains(t, outputInfo, testInfoMsg)

		log.Debug(testDebugMsg)
		outputDebug := sink.String()
		assert.Contains(t, outputDebug, testDebugMsg)
		sink.Reset()
	})

	t.Run("test info level", func(t *testing.T) {
		log := New("info", loggerCfg)

		log.Debug(testDebugMsg)
		outputDebug := sink.String()
		assert.NotContains(t, outputDebug, testDebugMsg)

		log.Info(testInfoMsg)
		outputInfo := sink.String()
		assert.Contains(t, outputInfo, testInfoMsg)
		sink.Reset()
	})

	t.Run("test warn level", func(t *testing.T) {
		log := New("warn", loggerCfg)

		log.Debug(testDebugMsg)
		outputDebug := sink.String()
		assert.NotContains(t, outputDebug, testDebugMsg)

		log.Warn(testWarnMsg)
		outputWarn := sink.String()
		assert.Contains(t, outputWarn, testWarnMsg)
		sink.Reset()
	})

	t.Run("test error level", func(t *testing.T) {
		log := New("error", loggerCfg)

		log.Info(testInfoMsg)
		outputInfo := sink.String()
		assert.NotContains(t, outputInfo, testInfoMsg)

		log.Error(testErrorMsg)
		outputError := sink.String()
		assert.Contains(t, outputError, testErrorMsg)
		sink.Reset()
	})

	t.Run("test fatal level", func(t *testing.T) {
		log := New("fatal", loggerCfg).WithOptions(zap.WithFatalHook(zapcore.WriteThenGoexit))

		// test example
		// from https://github.com/uber-go/zap/blob/d27427d23f81dba1f048d6034d5f286572049e1e/logger_test.go#L839
		var finished bool
		recovered := make(chan interface{})
		go func() {
			defer func() {
				recovered <- recover()
			}()

			log.Info(testInfoMsg)
			outputInfo := sink.String()
			assert.NotContains(t, outputInfo, testInfoMsg)

			//nolint:revive // test code
			log.Fatal(testFatalMsg)
			finished = true
		}()

		assert.Equal(t, nil, <-recovered, "unexpected value from recover()")
		assert.False(t, finished, "expect goroutine to not finish after Fatal")

		outputFatal := sink.String()
		assert.Contains(t, outputFatal, testFatalMsg)
	})

	t.Run("test default level", func(t *testing.T) {
		log := New("", loggerCfg)

		log.Info(testInfoMsg)
		outputInfo := sink.String()
		assert.Contains(t, outputInfo, testInfoMsg)

		log.Error(testErrorMsg)
		outputError := sink.String()
		assert.Contains(t, outputError, testErrorMsg)
		sink.Reset()
	})
}

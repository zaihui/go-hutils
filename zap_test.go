package hutils

import (
	"context"
	"errors"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap/zapcore"

	"github.com/stretchr/testify/assert"
)

func TestStdoutLog(t *testing.T) {
	output, err := CaptureStdout(func() {
		logger := &Logger{}
		sugarLog := logger.Init(LoggerOpt{EnableStdout: true}).Sugar()
		AccessLog{}.Log(sugarLog)
		RequestLog{}.Log(sugarLog)
		Error(sugarLog, errors.New("Error"))
		Track(sugarLog, "Track")
	})
	assert.Equal(t, err, nil)
	// 最开始会输出一条空行
	assert.Equal(t, len(output), 5)
}

func TestFileLog(t *testing.T) {
	// 生成临时文件夹
	path, err := ioutil.TempDir("", "logs")
	assert.Equal(t, err, nil)
	defer os.RemoveAll(path)
	logger := &Logger{LogPath: path, Type: TRACK}
	sugarLog := logger.Init(LoggerOpt{EnableFile: true}).Sugar()
	Track(sugarLog, "Track")

	file := path + "/track.log"
	_, err = os.Stat(file)
	assert.Equal(t, err, nil)
}

func TestCustomeEncoderConfig(t *testing.T) {
	output, err := CaptureStdout(func() {
		logger := &Logger{Type: ERROR}
		sugarLog := logger.Init(LoggerOpt{EnableStdout: true}).Sugar()
		sugarLog.Error("test")
	})
	assert.Equal(t, err, nil)
	// service_name level date time path func message
	assert.Equal(t, strings.Split(output[0], " ")[6], "test")

	output, err = CaptureStdout(func() {
		logger := &Logger{Type: ERROR}
		sugarLog := logger.Init(LoggerOpt{
			EnableStdout: true,
			CustomEncoderConfig: &zapcore.EncoderConfig{
				LevelKey:         "level",
				MessageKey:       "message",
				EncodeLevel:      zapcore.CapitalLevelEncoder,
				ConsoleSeparator: "\t",
			},
		}).Sugar()
		sugarLog.Info("test")
	})
	assert.Equal(t, err, nil)
	assert.Equal(t, output[0], "INFO\ttest")
}

func TestSetServiceName(t *testing.T) {
	assert.Equal(t, serviceName, "default")
	name := "golang"
	SetServiceName(name)
	assert.Equal(t, name, serviceName)
}

func TestStdoutLogWithContext(t *testing.T) {
	traceID, err := trace.TraceIDFromHex("744ba40615ac6737263c10f1255eac36")
	assert.NoError(t, err)
	spanID, err := trace.SpanIDFromHex("a221978841e89dac")
	assert.NoError(t, err)
	spanCtx := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: traceID,
		SpanID:  spanID,
	})
	ctx := trace.ContextWithSpanContext(context.Background(), spanCtx)
	output, err := CaptureStdout(func() {
		logger := &Logger{}
		sugarLog := logger.Init(LoggerOpt{EnableStdout: true}).Sugar()
		AccessLog{}.LogWithContext(ctx, sugarLog)
		RequestLog{}.LogWithContext(context.Background(), sugarLog)
	})
	assert.Equal(t, err, nil)
	assert.Equal(t, len(output), 3)
	assert.True(t, strings.Index(output[0], traceID.String()) > 0)
	assert.True(t, strings.Index(output[0], spanID.String()) > 0)
	assert.True(t, strings.Index(output[1], traceID.String()) == -1)
	assert.True(t, strings.Index(output[1], spanID.String()) == -1)
}

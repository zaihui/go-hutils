package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	rotateLogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LogType string

const (
	ACCESS  LogType = "access"
	REQUEST LogType = "request"
	TRACK   LogType = "track"
	ERROR   LogType = "error"
)

type Logger struct {
	Type    LogType
	LogPath string
}

type LoggerOpt struct {
	EnableStdout bool
	EnableFile   bool
}

func (l *Logger) Init(opt *LoggerOpt) (logger *zap.Logger) {
	writers := []zapcore.WriteSyncer{}
	if opt.EnableStdout {
		writers = append(writers, zapcore.AddSync(os.Stdout))
	}
	if opt.EnableFile {
		writers = append(writers, zapcore.AddSync(l.fileRotateWriter()))
	}
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(l.encoderConfig()),
		zapcore.NewMultiWriteSyncer(writers...),
		zapcore.InfoLevel,
	)
	return zap.New(core)
}

func (l *Logger) encoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		CallerKey:      "file",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder,
	}
}

func (l *Logger) fileRotateWriter() io.Writer {
	filePath := l.filePath()
	hook, err := rotateLogs.New(
		filePath+".%Y-%m-%d",
		rotateLogs.WithLinkName(filePath),
		rotateLogs.WithMaxAge(time.Hour*24*30),
		rotateLogs.WithRotationTime(time.Hour*24),
	)
	if err != nil {
		log.Panic(err)
	}
	return hook
}

func (l *Logger) filePath() string {
	return fmt.Sprintf("%s/%s.log", l.LogPath, l.Type)
}

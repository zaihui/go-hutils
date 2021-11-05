package hutils

import (
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	rotateLogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var serviceName = GetEnv("SERVICE_NAME", "default")

func SetServiceName(name string) {
	serviceName = name
}

const (
	timeFormatter  = "2006-01-02 15:04:05"
	defaultLogType = "http"
	grpcLogType    = "grpc"
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

type AccessLog struct {
	ClientIP   string
	Method     string
	Request    string
	Protocol   string
	Agent      string
	LogType    string
	GrpcStatus string
	Payload    []byte
	Response   []byte
	Duration   int64
	StatusCode int
}

func (l AccessLog) Log(logger *zap.SugaredLogger) {
	now := time.Now().Format(timeFormatter)
	logType := defaultLogType
	if l.LogType != "" {
		logType = l.LogType
	}
	logger.Infof(
		"%s %s %s %s $%q$ %s %d %d \"%s\" %s $%q$ %s %s",
		now, l.ClientIP, l.Method, l.Request, l.Payload, l.Protocol,
		l.StatusCode, l.Duration, l.Agent, serviceName, l.Response, logType, l.GrpcStatus,
	)
}

type RequestLog struct {
	Method            string
	Request           string
	StatusDescription string
	Payload           []byte
	Response          []byte
	Duration          int64
}

func (l RequestLog) Log(logger *zap.SugaredLogger) {
	now := time.Now().Format(timeFormatter)
	statusDescriptions := strings.Split(l.StatusDescription, " ")
	if len(statusDescriptions) > 2 {
		statusCode, err := strconv.Atoi(statusDescriptions[0])
		if err != nil {
			statusDescription := strings.Join(statusDescriptions, "")
			logger.Infof(
				"%s %s %d %s $%q$ %d %s $%q$ %s",
				now, l.Method, l.Duration, l.Request, l.Payload, statusCode, statusDescription, l.Response, serviceName,
			)
			return
		}
	}
	logger.Infof(
		"%s %s %d %s $%q$ %s $%q$ %s",
		now, l.Method, l.Duration, l.Request, l.Payload, l.StatusDescription, l.Response, serviceName,
	)
}

func Error(logger *zap.SugaredLogger, err error) {
	now := time.Now().Format(timeFormatter)
	logger.Errorf("%s ERROR %s %+v", serviceName, now, err)
}

func Track(logger *zap.SugaredLogger, message interface{}) {
	now := time.Now().Format(timeFormatter)
	logger.Infof("%s INFO %s %s", serviceName, now, message)
}

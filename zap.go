package hutils

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"go.opentelemetry.io/otel/trace"

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

func SpanIDFromContext(ctx context.Context) string {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.HasSpanID() {
		return spanCtx.SpanID().String()
	}
	return ""
}

func TraceIDFromContext(ctx context.Context) string {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.HasTraceID() {
		return spanCtx.TraceID().String()
	}
	return ""
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

func (l AccessLog) LogWithContext(ctx context.Context, logger *zap.SugaredLogger) {
	now := time.Now().Format(timeFormatter)
	logType := defaultLogType
	if l.LogType != "" {
		logType = l.LogType
	}
	logger.Infof(
		"%s %s %s %s $%q$ %s %d %d \"%s\" %s $%q$ %s %s %s %s",
		now, l.ClientIP, l.Method, l.Request, l.Payload, l.Protocol,
		l.StatusCode, l.Duration, l.Agent, serviceName, l.Response, logType, l.GrpcStatus,
		TraceIDFromContext(ctx), SpanIDFromContext(ctx),
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
	logger.Infof(
		"%s %s %d %s $%q$ %s $%q$ %s",
		now, l.Method, l.Duration, l.Request, l.Payload, l.StatusDescription, l.Response, serviceName,
	)
}

func (l RequestLog) LogWithContext(ctx context.Context, logger *zap.SugaredLogger) {
	now := time.Now().Format(timeFormatter)
	logger.Infof(
		"%s %s %d %s $%q$ %s $%q$ %s %s %s",
		now, l.Method, l.Duration, l.Request, l.Payload, l.StatusDescription, l.Response, serviceName,
		TraceIDFromContext(ctx), SpanIDFromContext(ctx),
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

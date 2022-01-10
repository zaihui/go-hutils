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
	CustomEncoderConfig *zapcore.EncoderConfig
	EnableStdout        bool
	EnableFile          bool
}

func (l *Logger) Init(opt LoggerOpt) (logger *zap.Logger) {
	writers := []zapcore.WriteSyncer{}
	if opt.EnableStdout {
		writers = append(writers, zapcore.AddSync(os.Stdout))
	}
	if opt.EnableFile {
		writers = append(writers, zapcore.AddSync(l.fileRotateWriter()))
	}
	var enc zapcore.EncoderConfig
	switch l.Type {
	case ERROR:
		enc = ErrorEncoderConfig()
	case TRACK:
		enc = TrackEncoderConfig()
	default:
		enc = DefaultEncoderConfig()
	}
	if opt.CustomEncoderConfig != nil {
		enc = *opt.CustomEncoderConfig
	}
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(enc),
		zapcore.NewMultiWriteSyncer(writers...),
		zap.NewAtomicLevelAt(zapcore.InfoLevel),
	)
	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
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

func DefaultEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:          "time",
		MessageKey:       "msg",
		ConsoleSeparator: " ",
		LineEnding:       zapcore.DefaultLineEnding,
		EncodeTime:       zapcore.TimeEncoderOfLayout(timeFormatter),
	}
}

func ErrorEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:          "time",
		CallerKey:        "path",
		FunctionKey:      "func",
		MessageKey:       "msg",
		StacktraceKey:    "stacktrace",
		ConsoleSeparator: " ",
		LineEnding:       zapcore.DefaultLineEnding,
		EncodeDuration:   zapcore.StringDurationEncoder,
		EncodeCaller:     zapcore.ShortCallerEncoder,
		EncodeLevel:      zapcore.CapitalLevelEncoder,
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(serviceName)
			enc.AppendString(zapcore.ErrorLevel.CapitalString())
			enc.AppendString(t.Format(timeFormatter))
		},
	}
}

func TrackEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:          "time",
		FunctionKey:      "func",
		MessageKey:       "msg",
		ConsoleSeparator: " ",
		LineEnding:       zapcore.DefaultLineEnding,
		EncodeDuration:   zapcore.StringDurationEncoder,
		EncodeLevel:      zapcore.CapitalLevelEncoder,
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(serviceName)
			enc.AppendString(zapcore.InfoLevel.CapitalString())
			enc.AppendString(t.Format(timeFormatter))
		},
	}
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
	logType := defaultLogType
	if l.LogType != "" {
		logType = l.LogType
	}
	logger.Infof(
		"%s %s %s $%q$ %s %d %d \"%s\" %s $%q$ %s %s",
		l.ClientIP, l.Method, l.Request, l.Payload, l.Protocol,
		l.StatusCode, l.Duration, l.Agent, serviceName, l.Response, logType, l.GrpcStatus,
	)
}

func (l AccessLog) LogWithContext(ctx context.Context, logger *zap.SugaredLogger) {
	logType := defaultLogType
	if l.LogType != "" {
		logType = l.LogType
	}
	logger.Infof(
		"%s %s %s $%q$ %s %d %d \"%s\" %s $%q$ %s %s %s %s",
		l.ClientIP, l.Method, l.Request, l.Payload, l.Protocol,
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
	logger.Infof(
		"%s %s %d %s $%q$ %s $%q$ %s",
		l.Method, l.Duration, l.Request, l.Payload, l.StatusDescription, l.Response, serviceName,
	)
}

func (l RequestLog) LogWithContext(ctx context.Context, logger *zap.SugaredLogger) {
	logger.Infof(
		"%s %d %s $%q$ %s $%q$ %s %s %s",
		l.Method, l.Duration, l.Request, l.Payload, l.StatusDescription, l.Response, serviceName,
		TraceIDFromContext(ctx), SpanIDFromContext(ctx),
	)
}

func Error(logger *zap.SugaredLogger, err error) {
	logger.Errorf("%v", err)
}

func Track(logger *zap.SugaredLogger, message interface{}) {
	logger.Infof("%s", message)
}

package hutils

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	rotateLogs "github.com/lestrrat-go/file-rotatelogs"
	"go.opentelemetry.io/otel/trace"
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
	UNION   LogType = "union"
)

type Logger struct {
	Type    LogType
	LogPath string
}

type LoggerOpt struct {
	CustomEncoderConfig  *zapcore.EncoderConfig
	CustomEncoderConfigs map[zapcore.Level]*zapcore.EncoderConfig
	EnableStdout         bool
	EnableFile           bool
	IsUnion              bool
	IsJSONEncoder        bool
}

func (l *Logger) Init(opt LoggerOpt) (logger *zap.Logger) {
	var writers []zapcore.WriteSyncer
	if opt.EnableStdout {
		writers = append(writers, zapcore.AddSync(os.Stdout))
	}
	if opt.EnableFile {
		writers = append(writers, zapcore.AddSync(l.fileRotateWriter()))
	}
	var core zapcore.Core
	if opt.IsUnion {
		encoderMap := make(map[zapcore.Level]zapcore.EncoderConfig)
		encoderMap[zapcore.ErrorLevel] = DefaultEncoderConfig(true)
		encoderMap[zapcore.InfoLevel] = DefaultEncoderConfig(false)
		encoderMap[zapcore.DebugLevel] = DefaultEncoderConfig(false)
		for level, encoder := range opt.CustomEncoderConfigs {
			encoderMap[level] = *encoder
		}
		cores := make([]zapcore.Core, len(encoderMap))
		i := 0
		for key, encoderConfig := range encoderMap {
			level := key
			encoder := zapcore.NewConsoleEncoder(encoderConfig)
			if opt.IsJSONEncoder {
				encoder = zapcore.NewJSONEncoder(encoderConfig)
			}
			cores[i] = zapcore.NewCore(
				encoder,
				zapcore.NewMultiWriteSyncer(writers...),
				zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
					return lev == level
				}),
			)
			i++
		}
		core = zapcore.NewTee(cores...)
	} else {
		var enc zapcore.EncoderConfig
		switch l.Type {
		case ERROR:
			enc = DefaultEncoderConfig(true)
		case TRACK:
			enc = TrackEncoderConfig()
		default:
			enc = DefaultEncoderConfig(false)
		}
		if opt.CustomEncoderConfig != nil {
			enc = *opt.CustomEncoderConfig
		}
		core = zapcore.NewCore(
			zapcore.NewConsoleEncoder(enc),
			zapcore.NewMultiWriteSyncer(writers...),
			zap.NewAtomicLevelAt(zapcore.InfoLevel),
		)
	}
	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel)).Named(serviceName)
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

func DefaultEncoderConfig(WithStack bool) zapcore.EncoderConfig {
	config := zapcore.EncoderConfig{
		NameKey:          "name",
		LevelKey:         "level",
		TimeKey:          "time",
		CallerKey:        "path",
		FunctionKey:      "func",
		MessageKey:       "msg",
		ConsoleSeparator: " ",
		LineEnding:       zapcore.DefaultLineEnding,
		EncodeCaller:     zapcore.ShortCallerEncoder,
		EncodeDuration:   zapcore.StringDurationEncoder,
		EncodeLevel:      zapcore.CapitalLevelEncoder,
		EncodeName:       zapcore.FullNameEncoder,
		EncodeTime:       zapcore.TimeEncoderOfLayout(timeFormatter),
	}
	if WithStack {
		config.StacktraceKey = "stacktrace"
		config.LineEnding = "$" + zapcore.DefaultLineEnding
	}
	return config
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

type GetExtraField func(ctx context.Context) string

type UnionLog struct {
	ExtraFields map[string]GetExtraField
	ClientIP    string
	Protocol    string
	Agent       string
	Method      string
	Request     string
	GrpcStatus  string
	LogType     string
	Payload     []byte
	Response    []byte
	Duration    int64
	StatusCode  int
}

func (l UnionLog) GetExtraFields(ctx context.Context, baseInfo []interface{}) []interface{} {
	values := make([]interface{}, len(l.ExtraFields))
	i := 0
	for k, f := range l.ExtraFields {
		values[i] = zap.String(k, f(ctx))
		i++
	}
	if baseInfo != nil {
		values = append(values, baseInfo...)
	}
	return values
}

func (l UnionLog) Log(ctx context.Context, logger *zap.SugaredLogger) {
	logType := defaultLogType
	if l.LogType != "" {
		logType = l.LogType
	}
	baseInfo := []interface{}{
		zap.String("client_ip", l.ClientIP),
		zap.String("protocol", l.Protocol),
		zap.String("agent", l.Agent),
		zap.String("method", l.Method),
		zap.String("request", l.Request),
		zap.ByteString("payload", l.Payload),
		zap.ByteString("response", l.Response),
		zap.Int64("duration", l.Duration),
		zap.Int("status_code", l.StatusCode),
		zap.String("log_type", logType),
		zap.String("grpc_status", l.GrpcStatus)}
	values := l.GetExtraFields(ctx, baseInfo)
	logger.Infow("", values...)
}

func (l UnionLog) Error(ctx context.Context, logger *zap.SugaredLogger, err error) {
	values := l.GetExtraFields(ctx, nil)
	logger.Errorw(err.Error(), values...)
}

func (l UnionLog) Track(ctx context.Context, logger *zap.SugaredLogger, msg interface{}) {
	values := l.GetExtraFields(ctx, nil)
	logger.Debugw(fmt.Sprintf("%+v", msg), values...)
}

func (l UnionLog) Errorf(ctx context.Context, logger *zap.SugaredLogger, template string, args ...interface{}) {
	l.Error(ctx, logger, fmt.Errorf(template, args...))
}

func (l UnionLog) Trackf(ctx context.Context, logger *zap.SugaredLogger, template string, args ...interface{}) {
	l.Track(ctx, logger, fmt.Sprintf(template, args...))
}

func Error(logger *zap.SugaredLogger, err error) {
	logger.Errorf("%v", err)
}

func Track(logger *zap.SugaredLogger, message interface{}) {
	logger.Infof("%s", message)
}

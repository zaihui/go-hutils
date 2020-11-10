package logging

import (
	"time"

	"github.com/zaihui/go-hutils/pkg/utils"
	"go.uber.org/zap"
)

const timeFormatter = "2006-01-02 15:04:05"

var serviceName = utils.GetEnv("SERVICE_NAME", "default")

type AccessLog struct {
	ClientIP   string
	Method     string
	Request    string
	Protocol   string
	Agent      string
	Duration   int64
	StatusCode int
	Payload    []byte
}

func (l AccessLog) Log(logger *zap.SugaredLogger) {
	now := time.Now().Format(timeFormatter)
	logger.Infof(
		"%s %s %s %s $%q$ %s %d %d \"%s\" %s",
		now, l.ClientIP, l.Method, l.Request, l.Payload, l.Protocol, l.StatusCode, l.Duration, l.Agent, serviceName,
	)
}

type RequestLog struct {
	Method            string
	Request           string
	StatusDescription string
	Duration          int64
	Payload           []byte
	Response          []byte
}

func (l RequestLog) Log(logger *zap.SugaredLogger) {
	now := time.Now().Format(timeFormatter)
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

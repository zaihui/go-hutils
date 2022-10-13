package hutils

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// JSONMarshal 类似json.Marshal(), 但不转义特殊符号
func JSONMarshal(v interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(v)
	return buffer.Bytes(), err
}

// NewUUID 生成string类型的uuid
func NewUUID() string {
	uid, _ := uuid.New().MarshalBinary()
	return hex.EncodeToString(uid)
}

// GetEnv 获取环境变量，不存在则使用默认值
func GetEnv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		value = fallback
	}
	return value
}

// CaptureStdout 获取func执行的标准输出
func CaptureStdout(f func()) ([]string, error) {
	r, w, _ := os.Pipe()
	// 替换原有os.Stdout
	stdout := os.Stdout
	os.Stdout = w

	f()

	var buf bytes.Buffer
	output := make(chan string, 1)
	errs := make(chan error, 1)

	go func() {
		_, err := io.Copy(&buf, r)
		output <- buf.String()
		errs <- err
		r.Close()
	}()

	os.Stdout = stdout
	w.Close()
	return strings.Split(<-output, "\n"), <-errs
}

type CodeError interface {
	Error() string
	ErrCode() string
	ErrMessage() string
}

// ZError
// nolint: govet // may be we need err stack
type ZError struct {
	Code    string
	Message string
	TraceID string
	SpanID  string
	Err     error
}

type ZErrorOption func(*ZError)

func WithError(err error) ZErrorOption {
	return func(z *ZError) {
		z.Err = errors.WithStack(err)
	}
}

func NewZError(ctx context.Context, code interface{}, message string, options ...ZErrorOption) *ZError {
	z := &ZError{
		Code:    fmt.Sprintf("%v", code),
		Message: message,
		TraceID: TraceIDFromContext(ctx),
		SpanID:  SpanIDFromContext(ctx),
	}
	for _, option := range options {
		option(z)
	}
	return z
}

func (z ZError) Error() string {
	return fmt.Sprintf("%s: %s", z.Code, z.Message)
}

func (z ZError) ErrCode() string {
	return z.Code
}

func (z ZError) ErrMessage() string {
	return z.Message
}

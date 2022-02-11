package hutils

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"os"

	"github.com/google/uuid"
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

type ZError struct {
	Code    string
	Message string
	TraceID string
	SpanID  string
}

func NewZError(ctx context.Context, code, message string) *ZError {
	return &ZError{
		Code:    code,
		Message: message,
		TraceID: TraceIDFromContext(ctx),
		SpanID:  SpanIDFromContext(ctx),
	}
}

func (z ZError) Error() string {
	return fmt.Sprintf("%s: %s", z.Code, z.Message)
}

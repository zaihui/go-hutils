package logging

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	grpc_testing "github.com/grpc-ecosystem/go-grpc-middleware/testing"
	pb_testproto "github.com/grpc-ecosystem/go-grpc-middleware/testing/testproto"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
)

var (
	goodPing = &pb_testproto.PingRequest{Value: "goodPing", SleepTimeMs: 999}
)

type LogTestSuite struct {
	*grpc_testing.InterceptorTestSuite
	reader, writer *os.File
}

func TestLogTestSuite(t *testing.T) {
	r, w, _ := os.Pipe()
	// 替换原有os.Stdout
	os.Stdout = w
	logger := &Logger{}
	sugarLog := logger.Init(&LoggerOpt{EnableStdout: true}).Sugar()
	s := &LogTestSuite{
		InterceptorTestSuite: &grpc_testing.InterceptorTestSuite{
			ServerOpts: []grpc.ServerOption{
				grpc.UnaryInterceptor(NewUnaryServerAccessLogInterceptor(sugarLog)),
			},
		},
	}
	s.reader = r
	s.writer = w
	suite.Run(t, s)
}

func (s *LogTestSuite) TestNewUnaryServerAccessLogInterceptor() {
	_, err := s.Client.Ping(s.SimpleCtx(), goodPing)
	s.NoError(err)

	var buf bytes.Buffer
	output := make(chan string, 1)
	go func() {
		io.Copy(&buf, s.reader)
		output <- buf.String()
		s.reader.Close()
	}()
	s.writer.Close()

	o := strings.Split(<-output, "\n")
	// 输出空行
	s.Len(o, 1)
	s.Equal(o[0], "")
}

func (s *LogTestSuite) TestMarshalJSON() {
	buf, err := MarshalJSON(goodPing)
	s.Nil(err)
	s.True(strings.Contains(string(buf), "goodPing"))
}

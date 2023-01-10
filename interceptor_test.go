package hutils

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky/reporter"
	grpc_testing "github.com/grpc-ecosystem/go-grpc-middleware/testing"
	pb_testproto "github.com/grpc-ecosystem/go-grpc-middleware/testing/testproto"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
)

var (
	goodPing      = &pb_testproto.PingRequest{Value: "goodPing", SleepTimeMs: 999}
	skyClientPing = &pb_testproto.PingRequest{Value: "skyClientPing", SleepTimeMs: 999}
	skyServerPing = &pb_testproto.PingRequest{Value: "skyServerPing", SleepTimeMs: 999}
)

type LogTestSuite struct {
	*grpc_testing.InterceptorTestSuite
	reader, writer *os.File
}

type SkywalkingClientTestSuite struct {
	*grpc_testing.InterceptorTestSuite
	reader, writer *os.File
}

type SkywalkingServerTestSuite struct {
	*grpc_testing.InterceptorTestSuite
	reader, writer *os.File
}

func TestLogTestSuite(t *testing.T) {
	r, w, _ := os.Pipe()
	// 替换原有os.Stdout
	os.Stdout = w
	logger := &Logger{}
	sugarLog := logger.Init(LoggerOpt{EnableStdout: true}).Sugar()
	s := &LogTestSuite{
		InterceptorTestSuite: &grpc_testing.InterceptorTestSuite{
			ServerOpts: []grpc.ServerOption{
				grpc.UnaryInterceptor(NewUnaryServerAccessLogInterceptor(sugarLog, nil)),
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

func TestSkywalkingClientTestSuite(t *testing.T) {
	r, w, _ := os.Pipe()
	// 替换原有os.Stdout
	report, err := reporter.NewLogReporter()
	if err != nil {
		return
	}
	tracer, err := go2sky.NewTracer("test", go2sky.WithReporter(report))
	os.Stdout = w
	s := &SkywalkingClientTestSuite{
		InterceptorTestSuite: &grpc_testing.InterceptorTestSuite{
			ClientOpts: []grpc.DialOption{
				grpc.WithUnaryInterceptor(NewUnaryClientSkywalkingInterceptor(tracer)),
			},
		},
	}
	s.reader = r
	s.writer = w
	suite.Run(t, s)
}

func (c *SkywalkingClientTestSuite) TestNewUnaryClientSkywalkingInterceptor() {
	_, err := c.Client.Ping(c.SimpleCtx(), skyClientPing)
	c.NoError(err)

	var buf bytes.Buffer
	output := make(chan string, 1)
	go func() {
		io.Copy(&buf, c.reader)
		output <- buf.String()
		c.reader.Close()
	}()
	c.writer.Close()

	o := strings.Split(<-output, "\n")
	// 输出空行
	c.Len(o, 1)
	c.Equal(o[0], "")
}

func TestSkywalkingServerTestSuite(t *testing.T) {
	r, w, _ := os.Pipe()
	// 替换原有os.Stdout
	os.Stdout = w
	report, err := reporter.NewLogReporter()
	if err != nil {
		return
	}
	tracer, err := go2sky.NewTracer("test", go2sky.WithReporter(report))
	s := &SkywalkingServerTestSuite{
		InterceptorTestSuite: &grpc_testing.InterceptorTestSuite{
			ServerOpts: []grpc.ServerOption{
				grpc.UnaryInterceptor(NewUnaryServerSkywalkingInterceptor(tracer)),
			},
		},
	}
	s.reader = r
	s.writer = w
	suite.Run(t, s)
}

func (s *SkywalkingServerTestSuite) TestNewUnaryServerSkywalkingInterceptor() {
	_, err := s.Client.Ping(s.SimpleCtx(), skyServerPing)
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

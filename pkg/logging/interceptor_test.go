package logging

import (
	"strings"
	"testing"

	grpc_testing "github.com/grpc-ecosystem/go-grpc-middleware/testing"
	pb_testproto "github.com/grpc-ecosystem/go-grpc-middleware/testing/testproto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/zaihui/go-hutils/pkg/utils"
	"google.golang.org/grpc"
)

var (
	goodPing = &pb_testproto.PingRequest{Value: "goodPing", SleepTimeMs: 999}
)

type LogTestSuite struct {
	*grpc_testing.InterceptorTestSuite
}

func TestLogTestSuite(t *testing.T) {
	logger := &Logger{}
	sugarLog := logger.Init(&LoggerOpt{EnableStdout: true}).Sugar()
	s := &LogTestSuite{
		InterceptorTestSuite: &grpc_testing.InterceptorTestSuite{
			ServerOpts: []grpc.ServerOption{
				grpc.UnaryInterceptor(NewUnaryServerAccessLogInterceptor(sugarLog)),
			},
		},
	}
	suite.Run(t, s)
}

func (s *LogTestSuite) TestNewUnaryServerAccessLogInterceptor() {
	output, err := utils.CaptureStdout(func() {
		_, err := s.Client.Ping(s.SimpleCtx(), goodPing)
		assert.Equal(s.T(), err, nil)
	})
	s.Nil(err)
	// 输出一行日志
	s.Equal(len(output), 1)
}

func (s *LogTestSuite) TestMarshalJSON() {
	buf, err := MarshalJSON(goodPing)
	s.Nil(err)
	s.True(strings.Contains(string(buf), "goodPing"))
}

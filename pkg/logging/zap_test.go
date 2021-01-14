package logging

import (
	"errors"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zaihui/go-hutils/pkg/utils"
)

func TestStdoutLog(t *testing.T) {
	output, err := utils.CaptureStdout(func() {
		logger := &Logger{}
		sugarLog := logger.Init(&LoggerOpt{EnableStdout: true}).Sugar()
		AccessLog{}.Log(sugarLog)
		RequestLog{}.Log(sugarLog)
		Error(sugarLog, errors.New("Error"))
		Track(sugarLog, "Track")
	})
	assert.Equal(t, err, nil)
	// 最开始会输出一条空行
	assert.Equal(t, len(output), 5)
}

func TestFileLog(t *testing.T) {
	// 生成临时文件夹
	path, err := ioutil.TempDir("", "logs")
	assert.Equal(t, err, nil)
	defer os.RemoveAll(path)
	logger := &Logger{LogPath: path, Type: TRACK}
	sugarLog := logger.Init(&LoggerOpt{EnableFile: true}).Sugar()
	Track(sugarLog, "Track")

	file := path + "/track.log"
	_, err = os.Stat(file)
	assert.Equal(t, err, nil)
}

func TestSetServiceName(t *testing.T) {
	assert.Equal(t, serviceName, "default")
	name := "golang"
	SetServiceName(name)
	assert.Equal(t, name, serviceName)
}

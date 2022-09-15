package hutils

import (
	"errors"
	"time"
)

const (
	// DateLayout 日期序列化.
	DateLayout = "2006-01-02"
	// DateTimeLayout 时间序列化
	DateTimeLayout = "2006-01-02T15:04:05"
)

// Time time.Date的快捷方法，省略sec，nsec，loc.
func Time(year int, month time.Month, day, hour, min int) time.Time {
	return time.Date(year, month, day, hour, min, 0, 0, time.Local)
}

// DefaultParseTime 默认时间解析.
func DefaultParseTime(value string) (time.Time, error) {
	return ParseTime(DateTimeLayout, value)
}

// ParseTime 当地时区解析时间字符串.
func ParseTime(layout string, value string) (time.Time, error) {
	return time.ParseInLocation(layout, value, time.Local)
}

// Period 时间区间.
type Period struct {
	Start, End time.Time
}

// ParseDateTimePeriod 解析时间区间.
func ParseDateTimePeriod(start, end string) (*Period, error) {
	if start == "" || end == "" {
		return nil, errors.New("查询区间不能为空")
	}
	startFrom, err := ParseTime(DateTimeLayout, start)
	if err != nil {
		return nil, errors.New("开始时间解析失败")
	}
	endTo, err := ParseTime(DateTimeLayout, end)
	if err != nil {
		return nil, errors.New("结束时间解析失败")
	}
	return &Period{Start: startFrom, End: endTo}, nil
}

// Tomorrow 明天同一时间.
func Tomorrow() time.Time {
	return time.Now().AddDate(0, 0, 1)
}

// Yesterday 昨天同一时间.
func Yesterday() time.Time {
	return time.Now().AddDate(0, 0, -1)
}

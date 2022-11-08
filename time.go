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
	// DateTimeLayoutWithoutT 没有T的时间序列化
	// Deprecated
	DateTimeLayoutWithoutT = ISO8601Local
	// DateTimeLayoutChinese 中文时间序列化
	DateTimeLayoutChinese = "2006年01月02日 15时04分05秒"

	// ISO8601Local 本地ISO8601日期序列化.
	ISO8601Local = "2006-01-02 15:04:05"
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

// GetDayStartAndLatest 获取某天的开始和结束.
func GetDayStartAndLatest(day time.Time) (time.Time, time.Time) {
	beginAt := Time(day.Year(), day.Month(), day.Day(), 0, 0)
	endAt := time.Date(day.Year(), day.Month(), day.Day(), 23, 59, 59, 59, time.Local)
	return beginAt, endAt
}

// DiffDay 两个时间差多少天.
func DiffDay(t1, t2 time.Time) int {
	var (
		begin time.Time
		end   time.Time
	)
	if t1.After(t2) {
		begin = t2
		end = t1
	} else {
		begin = t1
		end = t2
	}
	return int(end.Sub(begin).Hours() / 24)
}

// DiffDayWithLayout 两个时间差多少天.
func DiffDayWithLayout(t1, t2 string, layout string) (int, error) {
	if t1 == "" || t2 == "" {
		return 0, errors.New("查询区间不能为空")
	}
	t1T, err := ParseTime(layout, t1)
	if err != nil {
		return 0, errors.New("时间解析失败")
	}
	t2T, err := ParseTime(DateTimeLayout, t2)
	if err != nil {
		return 0, errors.New("时间解析失败")
	}
	return DiffDay(t1T, t2T), nil
}

package hutils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseTime(t *testing.T) {
	datetimeStr := "2021-09-30T12:00:00"
	datetime, err := DefaultParseTime(datetimeStr)
	assert.Nil(t, err)
	assert.Equal(t, time.Duration(0), datetime.Sub(Time(2021, 9, 30, 12, 0)))

	dateStr := "2021-09-30"
	date, err := ParseTime(DateLayout, dateStr)
	assert.Nil(t, err)
	assert.Equal(t, time.Duration(0), date.Sub(Time(2021, 9, 30, 0, 0)))
}

func TestParseTimePeriod(t *testing.T) {
	var start, end string
	_, err := ParseDateTimePeriod(start, end)
	assert.NotNil(t, err)
	start = "2021-09-30T12:00:00"
	end = "2021-10-07T20:00:00"
	period, err := ParseDateTimePeriod(start, end)
	assert.Nil(t, err)
	assert.Equal(t, Time(2021, 9, 30, 12, 0), period.Start)
	assert.Equal(t, Time(2021, 10, 7, 20, 0), period.End)
}

func TestGetDayStartAndLatest(t *testing.T) {
	now := time.Now()
	start, end := GetDayStartAndLatest(now)
	assert.Equal(t, Time(now.Year(), now.Month(), now.Day(), 0, 0), start)
	assert.Equal(t, time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 59, time.Local), end)
}

func TestDiffDay(t *testing.T) {
	begin := time.Now()
	end := begin.AddDate(0, 0, 3)
	diffDay := DiffDay(begin, end)
	assert.Equal(t, 3, diffDay)
	diffDay = DiffDay(end, begin)
	assert.Equal(t, 3, diffDay)
	diffDay = DiffDay(begin, begin)
	assert.Equal(t, 0, diffDay)
}

func TestDiffDayWithLayout(t *testing.T) {
	begin := time.Now()
	end := begin.AddDate(0, 0, 3)
	diffDay, err := DiffDayWithLayout(begin.Format(DateTimeLayout), end.Format(DateTimeLayout), DateTimeLayout)
	assert.Nil(t, err)
	assert.Equal(t, 3, diffDay)
	_, err = DiffDayWithLayout(begin.Format(DateTimeLayout), end.Format(DateTimeLayout), DateLayout)
	assert.Error(t, err)
}

func TestISO8601Local(t *testing.T) {
	datetime, err := ParseTime(ISO8601Local, "2022-01-01 10:00:00")
	assert.NoError(t, err)
	assert.Equal(t, Time(2022, 1, 1, 10, 0), datetime)
}

func TestDateTimeLayoutChinese(t *testing.T) {
	datetime, err := ParseTime(DateTimeLayoutChinese, "2022年01月01日 10时00分00秒")
	assert.NoError(t, err)
	assert.Equal(t, Time(2022, 1, 1, 10, 0), datetime)
}

func TestShortcut(t *testing.T) {
	_ = Tomorrow()
	_ = Yesterday()
}

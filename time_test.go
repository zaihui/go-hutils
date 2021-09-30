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

func TestShortcut(t *testing.T) {
	_ = Tomorrow()
	_ = Yesterday()
}

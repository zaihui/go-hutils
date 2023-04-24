package hutils

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestRetryFailed(t *testing.T) {
	var number int
	increaseNumber := func() error {
		number++
		return errors.New("error occurs")
	}

	err := Retry(increaseNumber, RDelay(time.Microsecond*50))

	assert.NotNil(t, err)
	assert.Equal(t, DefaultRTimes, number)

	number = 0
	err = Retry(increaseNumber, RDelay(time.Microsecond*50), RTimes(2))

	assert.NotNil(t, err)
	assert.Equal(t, 2, number)
}

func TestRetrySucceeded(t *testing.T) {
	var number int
	increaseNumber := func() error {
		number++
		if number == DefaultRTimes {
			return nil
		}
		return errors.New("error occurs")
	}

	err := Retry(increaseNumber, RDelay(time.Microsecond*50))

	assert.Nil(t, err)
	assert.Equal(t, DefaultRTimes, number)

	number = 0
	setTimeIncreaseNumber := func() error {
		number++
		if number == 2 {
			return nil
		}
		return errors.New("error occurs")
	}
	err = Retry(setTimeIncreaseNumber, RDelay(time.Microsecond*50), RTimes(2))

	assert.Nil(t, err)
	assert.Equal(t, 2, number)
}

func TestRetryCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	var number int
	increaseNumber := func() error {
		number++
		if number > 3 {
			cancel()
		}
		return errors.New("error occurs")
	}

	err := Retry(increaseNumber,
		RDelay(time.Microsecond*50),
		Context(ctx),
	)

	assert.NotNil(t, err)
	assert.Equal(t, 4, number)
}

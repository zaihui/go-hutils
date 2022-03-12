package hutils

import (
	"context"
	"errors"
	"time"
)

const (
	DefaultRTimes = 5
	DefaultRDelay = 3000 * time.Millisecond
)

type RetryConfig struct {
	context  context.Context
	times    uint
	duration time.Duration
}

type RetryFunc func() error

type Option func(*RetryConfig)

func RTimes(n uint) Option {
	return func(retryConfig *RetryConfig) {
		retryConfig.times = n
	}
}

func RDelay(d time.Duration) Option {
	return func(retryConfig *RetryConfig) {
		retryConfig.duration = d
	}
}

func Context(ctx context.Context) Option {
	return func(rc *RetryConfig) {
		rc.context = ctx
	}
}

func Retry(retryFunc RetryFunc, opt ...Option) error {
	retryConfig := &RetryConfig{
		context:  context.Background(),
		times:    DefaultRTimes,
		duration: DefaultRDelay,
	}
	for _, o := range opt {
		o(retryConfig)
	}
	for i := uint(0); i < retryConfig.times; i++ {
		err := retryFunc()
		if err != nil {
			select {
			case <-time.After(retryConfig.duration):
			case <-retryConfig.context.Done():
				return errors.New("retry is cancelled")
			}
		} else {
			return nil
		}
	}
	return errors.New("function run out of retry times")
}

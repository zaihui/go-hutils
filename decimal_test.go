package hutils

import (
	"errors"
	"testing"

	"github.com/shopspring/decimal"
)

func TestParseDecimal(t *testing.T) {
	testCases := []struct {
		input       interface{}
		expected    decimal.Decimal
		expectedErr error
	}{
		{123, decimal.NewFromInt(123), nil},
		{int8(123), decimal.NewFromInt(123), nil},
		{int16(123), decimal.NewFromInt(123), nil},
		{int32(123), decimal.NewFromInt(123), nil},
		{int64(123), decimal.NewFromInt(123), nil},
		{uint(123), decimal.NewFromInt(123), nil},
		{uint8(123), decimal.NewFromInt(123), nil},
		{uint16(123), decimal.NewFromInt(123), nil},
		{uint32(123), decimal.NewFromInt(123), nil},
		{uint64(123), decimal.NewFromInt(123), nil},
		{float32(123.45), decimal.NewFromFloat(123.45), nil},
		{123.456, decimal.NewFromFloat(123.456), nil},
		{"123.4567", decimal.NewFromFloat(123.4567), nil},
		{decimal.NewFromFloat(123.45678), decimal.NewFromFloat(123.45678), nil},
		{"not a number", decimal.Zero, errors.New("can't convert not a number to decimal: exponent is not numeric")},
		{struct{}{}, decimal.Zero, errors.New("unsupported data type")},
	}
	for _, tc := range testCases {
		actual, actualErr := ParseDecimal(tc.input)
		if actual.Cmp(tc.expected) != 0 {
			t.Errorf("ParseDecimal(%v) = %v, expected %v", tc.input, actual, tc.expected)
		}
		if (actualErr == nil && tc.expectedErr != nil) || (actualErr != nil && tc.expectedErr == nil) || (actualErr != nil && tc.expectedErr != nil && actualErr.Error() != tc.expectedErr.Error()) {
			t.Errorf("ParseDecimal(%v) error = %v, expected %v", tc.input, actualErr, tc.expectedErr)
		}
	}
}

package hutils

import (
	"errors"
	"math"
	"reflect"

	"github.com/shopspring/decimal"
)

func ParseDecimal(str interface{}) (decimal.Decimal, error) {
	var res decimal.Decimal
	var err error
	val := reflect.ValueOf(str)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	switch val.Interface().(type) {
	case int, int8, int16, int32, int64:
		res = decimal.NewFromInt(val.Int())
	case uint, uint8, uint16, uint32, uint64:
		v := val.Uint()
		if v > math.MaxInt64 {
			err = errors.New("uint value out of int64 range")
		} else {
			res = decimal.NewFromInt(int64(v))
		}
	case float32:
		res = decimal.NewFromFloat32(float32(val.Float()))
	case float64:
		res = decimal.NewFromFloat(val.Float())
	case string:
		res, err = decimal.NewFromString(val.String())
	case decimal.Decimal:
		res = val.Interface().(decimal.Decimal)
	default:
		err = errors.New("unsupported data type")
	}
	return res, err
}

package result

import (
	"fmt"
	"math/big"
	"time"
)

func (r *Rows) String(
	column string,
) (string, error) {

	v, err := r.Value(column)
	if err != nil {
		return "", err
	}

	x, ok := v.(string)
	if !ok {
		return "", typeError(column, "string", v)
	}

	return x, nil

}

func (r *Rows) Bytes(
	column string,
) ([]byte, error) {

	v, err := r.Value(column)
	if err != nil {
		return nil, err
	}

	x, ok := v.([]byte)
	if !ok {
		return nil, typeError(column, "[]byte", v)
	}

	return x, nil

}

func (r *Rows) Char(
	column string,
) (string, error) {

	return r.String(column)

}

func (r *Rows) Bool(
	column string,
) (bool, error) {

	v, err := r.Value(column)
	if err != nil {
		return false, err
	}

	x, ok := v.(bool)
	if !ok {
		return false, typeError(column, "bool", v)
	}

	return x, nil

}

func (r *Rows) Byte(
	column string,
) (int8, error) {

	v, err := r.Value(column)
	if err != nil {
		return 0, err
	}

	x, ok := v.(int8)
	if !ok {
		return 0, typeError(column, "int8", v)
	}

	return x, nil

}

func (r *Rows) Short(
	column string,
) (int16, error) {

	v, err := r.Value(column)
	if err != nil {
		return 0, err
	}

	x, ok := v.(int16)
	if !ok {
		return 0, typeError(column, "int16", v)
	}

	return x, nil

}

func (r *Rows) Integer(
	column string,
) (int32, error) {

	v, err := r.Value(column)
	if err != nil {
		return 0, err
	}

	x, ok := v.(int32)
	if !ok {
		return 0, typeError(column, "int32", v)
	}

	return x, nil

}

func (r *Rows) Long(
	column string,
) (int64, error) {

	v, err := r.Value(column)
	if err != nil {
		return 0, err
	}

	x, ok := v.(int64)
	if !ok {
		return 0, typeError(column, "int64", v)
	}

	return x, nil

}

func (r *Rows) BigInteger(
	column string,
) (*big.Int, error) {

	v, err := r.Value(column)
	if err != nil {
		return nil, err
	}

	x, ok := v.(*big.Int)
	if !ok {
		return nil, typeError(column, "*big.Int", v)
	}

	return x, nil

}

func (r *Rows) Float(
	column string,
) (float32, error) {

	v, err := r.Value(column)
	if err != nil {
		return 0, err
	}

	x, ok := v.(float32)
	if !ok {
		return 0, typeError(column, "float32", v)
	}

	return x, nil

}

func (r *Rows) Double(
	column string,
) (float64, error) {

	v, err := r.Value(column)
	if err != nil {
		return 0, err
	}

	x, ok := v.(float64)
	if !ok {
		return 0, typeError(column, "float64", v)
	}

	return x, nil

}

func (r *Rows) Decimal(
	column string,
) (*big.Float, error) {

	v, err := r.Value(column)
	if err != nil {
		return nil, err
	}

	x, ok := v.(*big.Float)
	if !ok {
		return nil, typeError(column, "*big.Float", v)
	}

	return x, nil

}

func (r *Rows) Date(
	column string,
) (time.Time, error) {

	v, err := r.Value(column)
	if err != nil {
		return time.Time{}, err
	}

	x, ok := v.(time.Time)
	if !ok {
		return time.Time{}, typeError(column, "time.Time", v)
	}

	return x, nil

}

func (r *Rows) Time(
	column string,
) (time.Time, error) {

	return r.Date(column)

}

func (r *Rows) Timestamp(
	column string,
) (time.Time, error) {

	return r.Date(column)

}

func (r *Rows) JSON(
	column string,
) (string, error) {

	return r.String(column)

}

func typeError(
	column string,
	expected string,
	actual interface{},
) error {

	return fmt.Errorf(
		"column %q contains %T instead of %s",
		column,
		actual,
		expected,
	)

}

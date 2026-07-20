package result

import (
	"fmt"
	"math/big"
	"time"

	kublingv1 "github.com/kubling-community/kubling-grpc/sdk-go/kubling/v1"
)

func DecodeValue(
	value *kublingv1.Value,
) (interface{}, error) {

	if value == nil {
		return nil, nil
	}

	switch v := value.GetKind().(type) {

	case *kublingv1.Value_NullValue:
		return nil, nil

	case *kublingv1.Value_StringValue:
		return v.StringValue, nil

	case *kublingv1.Value_VarbinaryValue:
		return v.VarbinaryValue, nil

	case *kublingv1.Value_CharValue:
		return v.CharValue, nil

	case *kublingv1.Value_BooleanValue:
		return v.BooleanValue, nil

	case *kublingv1.Value_ByteValue:
		return int8(v.ByteValue), nil

	case *kublingv1.Value_ShortValue:
		return int16(v.ShortValue), nil

	case *kublingv1.Value_IntegerValue:
		return v.IntegerValue, nil

	case *kublingv1.Value_LongValue:
		return v.LongValue, nil

	case *kublingv1.Value_BigintegerValue:
		i := new(big.Int)
		if _, ok := i.SetString(v.BigintegerValue, 10); !ok {
			return nil, fmt.Errorf("invalid biginteger value: %s", v.BigintegerValue)
		}
		return i, nil

	case *kublingv1.Value_FloatValue:
		return v.FloatValue, nil

	case *kublingv1.Value_DoubleValue:
		return v.DoubleValue, nil

	case *kublingv1.Value_BigdecimalValue:
		f, _, err := big.ParseFloat(v.BigdecimalValue, 10, 256, big.ToNearestEven)
		if err != nil {
			return nil, err
		}
		return f, nil

	case *kublingv1.Value_DateValue:
		return time.Parse("2006-01-02", v.DateValue)

	case *kublingv1.Value_TimeValue:
		return decodeTime(v.TimeValue)

	case *kublingv1.Value_TimestampValue:
		return decodeTimestamp(v.TimestampValue)

	case *kublingv1.Value_BlobValue:
		return v.BlobValue.GetData(), nil

	case *kublingv1.Value_ClobValue:
		return v.ClobValue.GetData(), nil

	case *kublingv1.Value_XmlValue:
		return v.XmlValue, nil

	case *kublingv1.Value_GeometryValue:
		return v.GeometryValue, nil

	case *kublingv1.Value_GeographyValue:
		return v.GeographyValue, nil

	case *kublingv1.Value_JsonValue:
		return v.JsonValue, nil

	default:
		return nil, fmt.Errorf("unsupported value kind %T", v)
	}

}

func decodeTimestamp(
	value string,
) (time.Time, error) {

	layouts := []string{
		"2006-01-02T15:04:05.999999999",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05",
	}

	for _, layout := range layouts {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid timestamp value: %s", value)
}

func decodeTime(
	value string,
) (time.Time, error) {

	layouts := []string{
		"15:04:05.999999999",
		"15:04:05",
	}

	for _, layout := range layouts {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid time value: %s", value)
}

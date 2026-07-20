package result

import (
	"testing"

	kublingv1 "github.com/kubling-community/kubling-grpc/sdk-go/kubling/v1"
)

func TestDecodeStringValue(t *testing.T) {

	value, err := DecodeValue(
		&kublingv1.Value{
			Kind: &kublingv1.Value_StringValue{
				StringValue: "hello",
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if value != "hello" {
		t.Fatalf(
			"expected hello, got %v",
			value,
		)
	}

}

func TestDecodeNullValue(t *testing.T) {

	value, err := DecodeValue(
		&kublingv1.Value{
			Kind: &kublingv1.Value_NullValue{
				NullValue: &kublingv1.NullValue{},
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if value != nil {
		t.Fatalf(
			"expected nil, got %v",
			value,
		)
	}

}

func TestDecodeIntegerValue(t *testing.T) {

	value, err := DecodeValue(
		&kublingv1.Value{
			Kind: &kublingv1.Value_IntegerValue{
				IntegerValue: 123,
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if value != int32(123) {
		t.Fatalf(
			"expected 123, got %v",
			value,
		)
	}

}

package result

import (
	"testing"

	kublingv1 "github.com/kubling-community/kubling-grpc/sdk-go/kubling/v1"
)

func TestString(t *testing.T) {

	r := New(
		[]*kublingv1.QueryBatch{
			{
				Columns: []*kublingv1.Column{
					{
						Name: "value",
					},
				},
				Rows: []*kublingv1.Row{
					{
						Values: []*kublingv1.Value{
							{
								Kind: &kublingv1.Value_StringValue{
									StringValue: "hello",
								},
							},
						},
					},
				},
			},
		},
	)

	rows := r.Rows()

	if !rows.Next() {
		t.Fatal("expected one row")
	}

	value, err := rows.String("value")
	if err != nil {
		t.Fatal(err)
	}

	if value != "hello" {
		t.Fatalf(
			"expected hello, got %q",
			value,
		)
	}

}

func TestInteger(t *testing.T) {

	r := New(
		[]*kublingv1.QueryBatch{
			{
				Columns: []*kublingv1.Column{
					{
						Name: "value",
					},
				},
				Rows: []*kublingv1.Row{
					{
						Values: []*kublingv1.Value{
							{
								Kind: &kublingv1.Value_IntegerValue{
									IntegerValue: 123,
								},
							},
						},
					},
				},
			},
		},
	)

	rows := r.Rows()

	if !rows.Next() {
		t.Fatal("expected one row")
	}

	value, err := rows.Integer("value")
	if err != nil {
		t.Fatal(err)
	}

	if value != 123 {
		t.Fatalf(
			"expected 123, got %d",
			value,
		)
	}

}

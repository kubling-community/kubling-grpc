package result

import (
	"testing"

	kublingv1 "github.com/kubling-community/kubling-grpc/sdk-go/kubling/v1"
)

func TestRowCount(t *testing.T) {

	r := New(
		[]*kublingv1.QueryBatch{
			{
				Rows: []*kublingv1.Row{
					{},
					{},
				},
			},
			{
				Rows: []*kublingv1.Row{
					{},
				},
			},
		},
	)

	if r.RowCount() != 3 {
		t.Fatalf(
			"expected 3 rows, got %d",
			r.RowCount(),
		)
	}

}

func TestColumnCount(t *testing.T) {

	r := New(
		[]*kublingv1.QueryBatch{
			{
				Columns: []*kublingv1.Column{
					{},
					{},
				},
			},
		},
	)

	if r.ColumnCount() != 2 {
		t.Fatalf(
			"expected 2 columns, got %d",
			r.ColumnCount(),
		)
	}

}

func TestRowsValue(t *testing.T) {

	r := New(
		[]*kublingv1.QueryBatch{
			{
				Columns: []*kublingv1.Column{
					{
						Name: "greeting",
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

	value, err := rows.Value("greeting")
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

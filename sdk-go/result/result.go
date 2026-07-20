package result

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/kubling-community/kubling-grpc/sdk-go/client"
	kublingv1 "github.com/kubling-community/kubling-grpc/sdk-go/kubling/v1"
)

type Result struct {
	batches       []*kublingv1.QueryBatch
	columnIndexes map[string]int
}

type Rows struct {
	result  *Result
	batch   int
	row     int
	current *kublingv1.Row
}

func New(
	batches []*kublingv1.QueryBatch,
) *Result {

	r := &Result{
		batches:       batches,
		columnIndexes: make(map[string]int),
	}

	if len(batches) > 0 {

		for i, c := range batches[0].GetColumns() {
			r.columnIndexes[c.GetName()] = i
		}

	}

	return r

}

func Query(
	client *client.Client,
	sql string,
) (*Result, error) {

	ctx, cancel :=
		context.WithTimeout(
			context.Background(),
			30*time.Second,
		)
	defer cancel()

	stream, err :=
		client.Query().Query(
			ctx,
			&kublingv1.QueryRequest{
				ExpiringToken: client.Token(),
				Sql:           sql,
			},
		)
	if err != nil {
		return nil, err
	}

	var batches []*kublingv1.QueryBatch

	for {

		batch, err := stream.Recv()

		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		batches = append(
			batches,
			batch,
		)

	}

	return New(batches), nil

}

func (r *Result) BatchCount() int {

	return len(r.batches)

}

func (r *Result) RowCount() int {

	count := 0

	for _, batch := range r.batches {

		count += len(batch.GetRows())

	}

	return count

}

func (r *Result) ColumnCount() int {

	if len(r.batches) == 0 {
		return 0
	}

	return len(
		r.batches[0].GetColumns(),
	)

}

func (r *Result) Rows() *Rows {

	return &Rows{
		result: r,
		batch:  0,
		row:    -1,
	}

}

func (r *Rows) Close() error {

	r.current = nil
	r.result = nil

	return nil

}

func (r *Rows) Next() bool {

	if r.result == nil {
		return false
	}

	for r.batch < len(r.result.batches) {

		batch := r.result.batches[r.batch]

		r.row++

		if r.row < len(batch.GetRows()) {

			r.current = batch.GetRows()[r.row]

			return true

		}

		r.batch++
		r.row = -1

	}

	r.current = nil

	return false

}

func (r *Rows) Value(
	column string,
) (interface{}, error) {

	if r.current == nil {

		return nil,
			fmt.Errorf(
				"there is no current row; call Next() first",
			)

	}

	index, ok :=
		r.result.columnIndexes[column]

	if !ok {

		return nil,
			fmt.Errorf(
				"unknown column %q",
				column,
			)

	}

	if index >= len(r.current.GetValues()) {

		return nil,
			fmt.Errorf(
				"column %q is out of bounds",
				column,
			)

	}

	return DecodeValue(
		r.current.GetValues()[index],
	)

}

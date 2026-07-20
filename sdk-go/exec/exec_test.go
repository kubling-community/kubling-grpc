package exec

import (
	"testing"

	kublingv1 "github.com/kubling-community/kubling-grpc/sdk-go/kubling/v1"
	"github.com/kubling-community/kubling-grpc/sdk-go/result"
)

func TestAffectedRows(t *testing.T) {

	execResult := &ExecResult{
		affectedRows: 10,
	}

	if execResult.AffectedRows() != 10 {
		t.Fatalf(
			"expected 10 rows, got %d",
			execResult.AffectedRows(),
		)
	}

}

func TestUpdateCounts(t *testing.T) {

	execResult := &ExecResult{
		updateCounts: []int64{
			1,
			2,
			3,
		},
	}

	counts := execResult.UpdateCounts()

	if len(counts) != 3 {
		t.Fatalf(
			"expected 3 counts, got %d",
			len(counts),
		)
	}

}

func TestGeneratedKeys(t *testing.T) {

	keys :=
		result.New(
			[]*kublingv1.QueryBatch{
				{
					Rows: []*kublingv1.Row{
						{},
					},
				},
			},
		)

	execResult := &ExecResult{
		generatedKeys: keys,
	}

	if execResult.GeneratedKeys() == nil {
		t.Fatal(
			"expected generated keys",
		)
	}

	if execResult.GeneratedKeys().RowCount() != 1 {
		t.Fatalf(
			"expected 1 key row, got %d",
			execResult.GeneratedKeys().RowCount(),
		)
	}

}

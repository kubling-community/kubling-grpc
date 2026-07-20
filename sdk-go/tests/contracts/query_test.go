package contracts

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	exec "github.com/kubling-community/kubling-grpc/sdk-go/exec"
	"github.com/kubling-community/kubling-grpc/sdk-go/internal/assert"
	res "github.com/kubling-community/kubling-grpc/sdk-go/result"
)

func TestQuery(
	t *testing.T,
) {

	client := runtime.Client(t)

	id := uuid.NewString()

	_, err := exec.Exec(
		client,
		fmt.Sprintf(
			`INSERT INTO TYPE_COVERAGE (
				ID,
				STRING_VALUE
			)
			VALUES (
				'%s',
				'hello world'
			)`,
			id,
		),
	)
	if err != nil {
		t.Fatal(err)
	}

	result, err := res.Query(
		client,
		fmt.Sprintf(
			`SELECT STRING_VALUE
			 FROM TYPE_COVERAGE
			 WHERE ID='%s'`,
			id,
		),
	)
	if err != nil {
		t.Fatal(err)
	}

	if err := assert.AssertEquals(
		1,
		result.ColumnCount(),
	); err != nil {
		t.Fatal(err)
	}

	if err := assert.AssertEquals(
		1,
		result.RowCount(),
	); err != nil {
		t.Fatal(err)
	}

	rows := result.Rows()

	if !rows.Next() {
		t.Fatal("expected one row")
	}

	value, err := rows.String("STRING_VALUE")
	if err != nil {
		t.Fatal(err)
	}

	if err := assert.AssertEquals(
		"hello world",
		value,
	); err != nil {
		t.Fatal(err)
	}

}

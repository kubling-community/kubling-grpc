package contracts

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	assert "github.com/kubling-community/kubling-grpc/sdk-go/internal/assert"
	res "github.com/kubling-community/kubling-grpc/sdk-go/result"
	tx "github.com/kubling-community/kubling-grpc/sdk-go/tx"
)

func TestTx(
	t *testing.T,
) {

	client := runtime.Client(t)

	tx, err := tx.Begin(client)
	if err != nil {
		t.Fatal(err)
	}

	defer tx.Rollback()

	id := uuid.NewString()

	//
	// INSERT
	//

	result, err := tx.Exec(
		fmt.Sprintf(
			`INSERT INTO TYPE_COVERAGE (
				ID,
				STRING_VALUE
			)
			VALUES (
				'%s',
				'before'
			)`,
			id,
		),
	)
	if err != nil {
		t.Fatal(err)
	}

	if err := assert.AssertEquals(
		int64(1),
		result.AffectedRows(),
	); err != nil {
		t.Fatal(err)
	}

	//
	// UPDATE
	//

	result, err = tx.Exec(
		fmt.Sprintf(
			`UPDATE TYPE_COVERAGE
			 SET STRING_VALUE='after'
			 WHERE ID='%s'`,
			id,
		),
	)
	if err != nil {
		t.Fatal(err)
	}

	if err := assert.AssertEquals(
		int64(1),
		result.AffectedRows(),
	); err != nil {
		t.Fatal(err)
	}

	//
	// VERIFY INSIDE TRANSACTION
	//

	queryResult, err := tx.Query(
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

	rows := queryResult.Rows()

	if !rows.Next() {
		t.Fatal("expected one row")
	}

	value, err := rows.String("STRING_VALUE")
	if err != nil {
		t.Fatal(err)
	}

	if err := assert.AssertEquals(
		"after",
		value,
	); err != nil {
		t.Fatal(err)
	}

	//
	// COMMIT
	//

	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}

	//
	// VERIFY AFTER COMMIT
	//

	queryResult, err = res.Query(
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

	rows = queryResult.Rows()

	if !rows.Next() {
		t.Fatal("expected one row")
	}

	value, err = rows.String("STRING_VALUE")
	if err != nil {
		t.Fatal(err)
	}

	if err := assert.AssertEquals(
		"after",
		value,
	); err != nil {
		t.Fatal(err)
	}

}

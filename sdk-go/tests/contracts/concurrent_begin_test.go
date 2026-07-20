package contracts

import (
	"context"
	"testing"
	"time"

	"github.com/kubling-community/kubling-grpc/sdk-go/client"
	"github.com/kubling-community/kubling-grpc/sdk-go/result"
	"github.com/kubling-community/kubling-grpc/sdk-go/tx"
)

func TestBeginOrGetReturnsExistingTransaction(
	t *testing.T,
) {

	clientA :=
		runtime.Client(
			t,
		)

	clientB, err :=
		client.NewClientWithToken(
			client.Options{
				Address: runtime.Address,
			},
			clientA.Token(),
		)
	if err != nil {
		t.Fatal(err)
	}
	defer clientB.Close()

	//
	// Begin transaction using client A.
	//

	transactionA, err :=
		tx.Begin(
			clientA,
		)
	if err != nil {
		t.Fatal(err)
	}
	defer transactionA.Rollback()

	//
	// Client B shares the same session, therefore
	// BeginOrGet() must return the existing transaction.
	//

	transactionB, err :=
		tx.BeginOrGet(
			clientB,
		)
	if err != nil {
		t.Fatal(err)
	}

	//
	// Execute through the second client.
	//

	_, err =
		transactionB.Exec(`
			INSERT INTO TYPE_COVERAGE (
				ID,
				STRING_VALUE
			)
			VALUES (
				'begin-or-get',
				'shared'
			)
		`)
	if err != nil {
		t.Fatal(err)
	}

	//
	// The first client must observe the row before commit,
	// proving both clients operate on the same transaction.
	//

	rows, err :=
		transactionA.Query(`
			SELECT STRING_VALUE
			FROM TYPE_COVERAGE
			WHERE ID='begin-or-get'
		`)
	if err != nil {
		t.Fatal(err)
	}

	if rows.RowCount() != 1 {
		t.Fatalf(
			"expected 1 row, got %d",
			rows.RowCount(),
		)
	}

	iterator := rows.Rows()

	if !iterator.Next() {
		t.Fatal("expected one row")
	}

	value, err := iterator.String("STRING_VALUE")
	if err != nil {
		t.Fatal(err)
	}

	if value != "shared" {
		t.Fatalf(
			"expected 'shared', got %#v",
			value,
		)
	}

	//
	// Commit using client A.
	//

	if err := transactionA.Commit(); err != nil {
		t.Fatal(err)
	}

	//
	// A new independent session must now observe the row.
	//

	clientC :=
		runtime.Client(
			t,
		)

	resultSet, err :=
		result.Query(
			clientC,
			`
			SELECT STRING_VALUE
			FROM TYPE_COVERAGE
			WHERE ID='begin-or-get'
			`,
		)
	if err != nil {
		t.Fatal(err)
	}

	if resultSet.RowCount() != 1 {
		t.Fatalf(
			"expected 1 row after commit, got %d",
			resultSet.RowCount(),
		)
	}

	iterator = resultSet.Rows()

	if !iterator.Next() {
		t.Fatal("expected one row")
	}

	value, err = iterator.String("STRING_VALUE")
	if err != nil {
		t.Fatal(err)
	}

	if value != "shared" {
		t.Fatalf(
			"expected 'shared', got %#v",
			value,
		)
	}
}

func TestSharedSessionDetectsActiveTransaction(
	t *testing.T,
) {

	clientA := runtime.Client(t)

	clientB, err :=
		client.NewClientWithToken(
			client.Options{
				Address: runtime.Address,
			},
			clientA.Token(),
		)
	if err != nil {
		t.Fatal(err)
	}
	defer clientB.Close()

	txA, err := tx.Begin(clientA)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = txA.Rollback()
	}()

	ctx, cancel :=
		context.WithTimeout(
			context.Background(),
			30*time.Second,
		)
	defer cancel()

	status, err :=
		clientB.IsInTransaction(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if !status.GetActive() {
		t.Fatal(
			"expected shared session to report an active transaction",
		)
	}

	txB, err := tx.BeginOrGet(clientB)
	if err != nil {
		t.Fatal(err)
	}

	if txB == nil {
		t.Fatal(
			"expected transaction handle for active shared session",
		)
	}
}

package contracts

import (
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/kubling-community/kubling-grpc/sdk-go/client"
	exec "github.com/kubling-community/kubling-grpc/sdk-go/exec"
	"github.com/kubling-community/kubling-grpc/sdk-go/internal/assert"
	res "github.com/kubling-community/kubling-grpc/sdk-go/result"
	"github.com/kubling-community/kubling-grpc/sdk-go/tx"
)

func TestCollaborativeTransaction(
	t *testing.T,
) {

	clientA := runtime.Client(t)

	clientB, err := client.NewClientWithToken(
		client.Options{
			Address: runtime.Address,
		},
		clientA.Token(),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer clientB.Close()

	clientC, err := client.NewClientWithToken(
		client.Options{
			Address: runtime.Address,
		},
		clientA.Token(),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer clientC.Close()

	transaction, err := tx.Begin(clientA)
	if err != nil {
		t.Fatal(err)
	}

	id1 := uuid.NewString()
	id2 := uuid.NewString()

	//
	// Client B participates in A's transaction.
	//

	_, err = exec.Exec(
		clientB,
		fmt.Sprintf(
			`INSERT INTO TYPE_COVERAGE (
				ID,
				STRING_VALUE
			)
			VALUES (
				'%s',
				'from-b'
			)`,
			id1,
		),
	)
	if err != nil {
		t.Fatal(err)
	}

	//
	// Client C also participates.
	//

	_, err = exec.Exec(
		clientC,
		fmt.Sprintf(
			`INSERT INTO TYPE_COVERAGE (
				ID,
				STRING_VALUE
			)
			VALUES (
				'%s',
				'from-c'
			)`,
			id2,
		),
	)
	if err != nil {
		t.Fatal(err)
	}

	//
	// Client A can already see both rows before commit.
	//

	result, err := res.Query(
		clientA,
		fmt.Sprintf(
			`SELECT COUNT(*) as count
			 FROM TYPE_COVERAGE
			 WHERE ID IN ('%s','%s')`,
			id1,
			id2,
		),
	)
	if err != nil {
		t.Fatal(err)
	}

	rows := result.Rows()

	if !rows.Next() {
		t.Fatal("expected one row")
	}

	count, err := rows.Integer("count")
	if err != nil {
		t.Fatal(err)
	}

	if err := assert.AssertEquals(
		int32(2),
		count,
	); err != nil {
		t.Fatal(err)
	}

	//
	// Commit performed only once.
	//

	if err := transaction.Commit(); err != nil {
		t.Fatal(err)
	}

}

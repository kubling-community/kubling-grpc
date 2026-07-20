package contracts

import (
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/kubling-community/kubling-grpc/sdk-go/client"
	"github.com/kubling-community/kubling-grpc/sdk-go/internal/assert"
	res "github.com/kubling-community/kubling-grpc/sdk-go/result"
	"github.com/kubling-community/kubling-grpc/sdk-go/tx"
)

func TestSharedSession(
	t *testing.T,
) {

	clientA :=
		runtime.Client(t)

	clientB, err :=
		client.NewClientWithToken(
			client.Options{
				Address:  runtime.Address,
				Username: "sa",
				Password: "sa",
				VDBName:  "TestVDB",
			},
			clientA.Token(),
		)
	if err != nil {
		t.Fatal(err)
	}
	defer clientB.Close()

	transaction, err :=
		tx.Begin(
			clientA,
		)
	if err != nil {
		t.Fatal(err)
	}

	id := uuid.NewString()

	_, err =
		transaction.Exec(
			fmt.Sprintf(
				`INSERT INTO TYPE_COVERAGE (
					ID,
					STRING_VALUE
				)
				VALUES (
					'%s',
					'shared-session'
				)`,
				id,
			),
		)
	if err != nil {
		t.Fatal(err)
	}

	result, err :=
		res.Query(
			clientB,
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

	rows := result.Rows()

	if !rows.Next() {
		t.Fatal("expected one row")
	}

	value, err := rows.String(
		"STRING_VALUE",
	)
	if err != nil {
		t.Fatal(err)
	}

	err = assert.AssertEquals(
		"shared-session",
		value,
	)
	if err != nil {
		t.Fatal(err)
	}

	err =
		transaction.Commit()
	if err != nil {
		t.Fatal(err)
	}

}

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

func TestRollbackSharedSession(
	t *testing.T,
) {

	clientA := runtime.Client(t)

	_, err := exec.Exec(
		clientA,
		"DELETE FROM TYPE_COVERAGE",
	)
	if err != nil {
		t.Fatal(err)
	}

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

	id := uuid.NewString()

	_, err = exec.Exec(
		clientB,
		fmt.Sprintf(`
			INSERT INTO TYPE_COVERAGE (
				ID,
				STRING_VALUE
			)
			VALUES (
				'%s',
				'rollback-test'
			)
		`, id),
	)
	if err != nil {
		t.Fatal(err)
	}

	//
	// Transaction must be active for every shared client.
	//

	for _, c := range []*client.Client{
		clientA,
		clientB,
		clientC,
	} {

		iit, err := c.IsInTransaction(
			t.Context(),
		)
		if err != nil {
			t.Fatal(err)
		}

		assert.AssertEquals(
			iit.Active,
			true,
		)

	}

	//
	// Shared session must observe the row before rollback.
	//

	for _, c := range map[string]*client.Client{
		"A": clientA,
		"B": clientB,
		"C": clientC,
	} {

		result, err := res.Query(
			c,
			fmt.Sprintf(`
				SELECT COUNT(*) AS total
				FROM TYPE_COVERAGE
				WHERE ID='%s'
			`, id),
		)
		if err != nil {
			t.Fatal(err)
		}

		rows := result.Rows()

		if !rows.Next() {
			t.Fatal("expected one row")
		}

		count, err := rows.Integer("total")
		if err != nil {
			t.Fatal(err)
		}

		assert.AssertEquals(
			count,
			int32(1),
		)

	}

	//
	// Rollback.
	//

	err = transaction.Rollback()
	if err != nil {
		t.Fatal(err)
	}

	//
	// Transaction must now be closed for the shared session.
	//

	for _, c := range []*client.Client{
		clientA,
		clientB,
		clientC,
	} {

		iit, err := c.IsInTransaction(
			t.Context(),
		)
		if err != nil {
			t.Fatal(err)
		}

		assert.AssertEquals(
			iit.Active,
			false,
		)

	}

	//
	// Every client should now observe zero rows.
	//

	clientD := runtime.Client(t)

	for _, c := range map[string]*client.Client{
		"A": clientA,
		"B": clientB,
		"C": clientC,
		"D": clientD,
	} {

		result, err := res.Query(
			c,
			fmt.Sprintf(`
		SELECT COUNT(*) AS total
		FROM TYPE_COVERAGE
		WHERE ID='%s'
	`, id),
		)
		if err != nil {
			t.Fatal(err)
		}

		rows := result.Rows()

		if !rows.Next() {
			t.Fatal("expected one row")
		}

		count, err := rows.Integer("total")
		if err != nil {
			t.Fatal(err)
		}

		assert.AssertEquals(
			count,
			int32(0),
		)

	}

}

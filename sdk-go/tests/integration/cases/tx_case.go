package cases

import (
	cli "github.com/kubling-community/kubling-grpc/sdk-go/client"
	assert "github.com/kubling-community/kubling-grpc/sdk-go/internal/assert"
	tx "github.com/kubling-community/kubling-grpc/sdk-go/tx"
)

func TxTest(
	client *cli.Client,
) error {

	tx, err :=
		tx.Begin(
			client,
		)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	result, err :=
		tx.Exec(
			"UPDATE portable_1.ACT_RE_DEPLOYMENT SET NAME_='tx-test' WHERE source_='source'",
		)
	if err != nil {
		return err
	}

	err = assert.AssertTrue(
		result.AffectedRows() > 0,
		"expected affected rows",
	)
	if err != nil {
		return err
	}

	return tx.Commit()

}

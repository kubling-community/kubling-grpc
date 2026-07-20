package cases

import (
	"fmt"

	cli "github.com/kubling-community/kubling-grpc/sdk-go/client"
	assert "github.com/kubling-community/kubling-grpc/sdk-go/internal/assert"
	res "github.com/kubling-community/kubling-grpc/sdk-go/result"
)

func QueryTest(
	client *cli.Client,
) error {

	result, err :=
		res.Query(
			client,
			"SELECT source_ FROM portable_1.ACT_RE_DEPLOYMENT LIMIT 1",
		)
	if err != nil {
		return err
	}

	err = assert.AssertEquals(
		1,
		result.ColumnCount(),
	)
	if err != nil {
		return err
	}

	err = assert.AssertEquals(
		1,
		result.RowCount(),
	)
	if err != nil {
		return err
	}

	rows := result.Rows()

	if !rows.Next() {
		return fmt.Errorf("expected one row")
	}

	source, err :=
		rows.String(
			"source_",
		)
	if err != nil {
		return err
	}

	return assert.AssertTrue(
		source != "",
		"source should not be empty",
	)

}

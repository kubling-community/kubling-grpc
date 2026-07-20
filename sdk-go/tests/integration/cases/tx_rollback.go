package cases

import (
	"context"
	"time"

	cli "github.com/kubling-community/kubling-grpc/sdk-go/client"
	assert "github.com/kubling-community/kubling-grpc/sdk-go/internal/assert"
	kublingv1 "github.com/kubling-community/kubling-grpc/sdk-go/kubling/v1"
)

func TxRollbackTest(
	client *cli.Client,
) error {

	ctx, cancel :=
		context.WithTimeout(
			context.Background(),
			30*time.Second,
		)
	defer cancel()

	_, err :=
		client.Query().BeginTransaction(
			ctx,
			&kublingv1.BeginTransactionRequest{
				ExpiringToken: client.Token(),
			},
		)
	if err != nil {
		return err
	}

	execResp, err :=
		client.Query().Exec(
			ctx,
			&kublingv1.ExecRequest{
				ExpiringToken: client.Token(),
				Sql:           "UPDATE portable_1.ACT_RE_DEPLOYMENT SET NAME_='rollback-test' WHERE source_='source';",
			},
		)
	if err != nil {
		return err
	}

	err = assert.AssertTrue(
		execResp.GetAffectedRows() > 0,
		"no rows updated",
	)
	if err != nil {
		return err
	}

	_, err =
		client.Query().RollbackTransaction(
			ctx,
			&kublingv1.RollbackTransactionRequest{
				ExpiringToken: client.Token(),
			},
		)

	return err

}

package cases

import (
	"context"
	"fmt"
	"time"

	cli "github.com/kubling-community/kubling-grpc/sdk-go/client"
	kublingv1 "github.com/kubling-community/kubling-grpc/sdk-go/kubling/v1"
)

func TxCommitTest(
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
				Sql:           "UPDATE portable_1.ACT_RE_DEPLOYMENT SET NAME_='modified from go' WHERE source_='source';",
			},
		)
	if err != nil {
		return err
	}

	fmt.Println(
		"affected rows:",
		execResp.GetAffectedRows(),
	)

	_, err =
		client.Query().CommitTransaction(
			ctx,
			&kublingv1.CommitTransactionRequest{
				ExpiringToken: client.Token(),
			},
		)

	return err

}

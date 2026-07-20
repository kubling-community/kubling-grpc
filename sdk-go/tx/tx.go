package tx

import (
	"context"
	"time"

	"github.com/kubling-community/kubling-grpc/sdk-go/client"
	"github.com/kubling-community/kubling-grpc/sdk-go/exec"
	res "github.com/kubling-community/kubling-grpc/sdk-go/result"

	kublingv1 "github.com/kubling-community/kubling-grpc/sdk-go/kubling/v1"
)

type Tx struct {
	client *client.Client
}

type Status struct {
	Active bool
}

func Begin(
	client *client.Client,
) (*Tx, error) {

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
		return nil, err
	}

	return FromClient(client), nil
}

func BeginOrGet(
	client *client.Client,
) (*Tx, error) {

	ctx, cancel :=
		context.WithTimeout(
			context.Background(),
			30*time.Second,
		)
	defer cancel()

	status, err :=
		client.IsInTransaction(ctx)
	if err != nil {
		return nil, err
	}

	if status.GetActive() {
		return FromClient(client), nil
	}

	return Begin(client)
}

func FromClient(
	client *client.Client,
) *Tx {

	return &Tx{
		client: client,
	}
}

func (tx *Tx) Commit() error {

	ctx, cancel :=
		context.WithTimeout(
			context.Background(),
			30*time.Second,
		)
	defer cancel()

	_, err :=
		tx.client.Query().CommitTransaction(
			ctx,
			&kublingv1.CommitTransactionRequest{
				ExpiringToken: tx.client.Token(),
			},
		)

	return err

}

func (tx *Tx) Rollback() error {

	ctx, cancel :=
		context.WithTimeout(
			context.Background(),
			30*time.Second,
		)
	defer cancel()

	_, err :=
		tx.client.Query().RollbackTransaction(
			ctx,
			&kublingv1.RollbackTransactionRequest{
				ExpiringToken: tx.client.Token(),
			},
		)

	return err

}

func (tx *Tx) Exec(
	sql string,
) (*exec.ExecResult, error) {

	return exec.Exec(
		tx.client,
		sql,
	)

}

func (tx *Tx) Query(
	sql string,
) (*res.Result, error) {

	return res.Query(
		tx.client,
		sql,
	)

}

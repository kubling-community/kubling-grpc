package exec

import (
	"context"
	"time"

	"github.com/kubling-community/kubling-grpc/sdk-go/client"
	kublingv1 "github.com/kubling-community/kubling-grpc/sdk-go/kubling/v1"
	res "github.com/kubling-community/kubling-grpc/sdk-go/result"
)

type ExecResult struct {
	affectedRows  int64
	updateCounts  []int64
	generatedKeys *res.Result
}

func Exec(
	client *client.Client,
	sql string,
) (*ExecResult, error) {

	ctx, cancel :=
		context.WithTimeout(
			context.Background(),
			30*time.Second,
		)
	defer cancel()

	response, err :=
		client.Query().Exec(
			ctx,
			&kublingv1.ExecRequest{
				ExpiringToken: client.Token(),
				Sql:           sql,
			},
		)
	if err != nil {
		return nil, err
	}

	result :=
		&ExecResult{
			affectedRows: response.GetAffectedRows(),
			updateCounts: response.GetUpdateCounts(),
		}

	if response.GetGeneratedKeys() != nil {

		result.generatedKeys =
			res.New(
				[]*kublingv1.QueryBatch{
					response.GetGeneratedKeys(),
				},
			)

	}

	return result, nil

}

func (r *ExecResult) AffectedRows() int64 {

	return r.affectedRows

}

func (r *ExecResult) UpdateCounts() []int64 {

	return r.updateCounts

}

func (r *ExecResult) GeneratedKeys() *res.Result {

	return r.generatedKeys

}

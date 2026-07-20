package cases

import (
	cli "github.com/kubling-community/kubling-grpc/sdk-go/client"
	assert "github.com/kubling-community/kubling-grpc/sdk-go/internal/assert"
	res "github.com/kubling-community/kubling-grpc/sdk-go/result"
)

func InvalidTokenCase(
	client *cli.Client,
) error {

	err := client.Logout()
	if err != nil {
		return err
	}

	_, err = res.Query(
		client,
		"SELECT 1",
	)

	return assert.AssertErrorContains(
		err,
		"Unauthenticated",
	)

}

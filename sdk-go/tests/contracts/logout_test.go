package contracts

import (
	"testing"

	assert "github.com/kubling-community/kubling-grpc/sdk-go/internal/assert"
	res "github.com/kubling-community/kubling-grpc/sdk-go/result"
)

func TestLogout(
	t *testing.T,
) {

	client :=
		runtime.Client(
			t,
		)

	err := client.Logout()
	if err != nil {
		t.Fatal(err)
	}

	_, err = res.Query(
		client,
		"SELECT 1",
	)

	assert.AssertErrorContains(
		err,
		"Unauthenticated",
	)

}

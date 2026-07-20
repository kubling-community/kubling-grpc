package support

import (
	"testing"

	cli "github.com/kubling-community/kubling-grpc/sdk-go/client"
)

func (r *Runtime) Client(
	t *testing.T,
) *cli.Client {

	t.Helper()

	client, err :=
		cli.NewClient(
			cli.Options{
				Address:  r.Address,
				Username: "sa",
				Password: "sa",
				VDBName:  "TestVDB",
			},
		)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {

		_ = client.Close()

	})

	return client

}

package main

import (
	"fmt"
	"log"

	cli "github.com/kubling-community/kubling-grpc/sdk-go/client"
	"github.com/kubling-community/kubling-grpc/sdk-go/tests/integration/cases"
)

type TestCase struct {
	Name string
	Run  func(*cli.Client) error
}

var tests = []TestCase{
	{
		Name: "tx-commit",
		Run:  cases.TxCommitTest,
	},
	{
		Name: "tx-rollback",
		Run:  cases.TxRollbackTest,
	},
	{
		Name: "query",
		Run:  cases.QueryTest,
	},
	{
		Name: "tx",
		Run:  cases.TxTest,
	},
	{
		Name: "logout",
		Run:  cases.InvalidTokenCase,
	},
}

func main() {

	client, err :=
		cli.NewClient(
			cli.Options{
				Address:  "localhost",
				Username: "sa",
				Password: "sa",
				VDBName:  "TestVDB",
			},
		)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	failures := 0

	for _, test := range tests {

		err := test.Run(client)

		if err != nil {

			failures++

			fmt.Printf(
				"FAIL %s: %v\n",
				test.Name,
				err,
			)

		} else {

			fmt.Printf(
				"PASS %s\n",
				test.Name,
			)

		}

	}

	fmt.Println()

	fmt.Printf(
		"%d tests, %d failures\n",
		len(tests),
		failures,
	)

}

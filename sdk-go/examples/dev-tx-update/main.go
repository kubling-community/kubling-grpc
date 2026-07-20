package main

import (
	"fmt"
	"log"

	"github.com/kubling-community/kubling-grpc/sdk-go/client"
	res "github.com/kubling-community/kubling-grpc/sdk-go/result"
)

func main() {

	cli, err := client.NewClient(
		client.Options{
			Address:  "localhost:50051",
			Username: "sa",
			Password: "sa",
			VDBName:  "KubernetesVDB",
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	result, err := res.Query(
		cli,
		`
		SELECT
			metadata__namespace,
			metadata__name
		FROM kube1.POD
		WHERE jsonPath(
			status__containerStatuses,
			'$[*].state.waiting.reason'
		) = CONVERT('["ImagePullBackOff"]', json)
		`,
	)
	if err != nil {
		log.Fatal(err)
	}

	rows := result.Rows()
	defer rows.Close()

	for rows.Next() {

		namespace, err := rows.String("metadata__namespace")
		if err != nil {
			log.Fatal(err)
		}

		name, err := rows.String("metadata__name")
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%s/%s\n", namespace, name)

	}

}

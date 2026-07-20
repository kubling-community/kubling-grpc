package support

import (
	"context"
	"fmt"
	"path/filepath"

	containertypes "github.com/moby/moby/api/types/container"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	KublingImage = "kubling/kubling:latest"

	MinimalScenario = "minimal"
)

type Runtime struct {
	container testcontainers.Container
	Address   string
}

type StdoutLogConsumer struct{}

func (c *StdoutLogConsumer) Accept(l testcontainers.Log) {
	fmt.Print(string(l.Content))
}

func StartRuntime(
	scenario string,
) (*Runtime, error) {

	ctx := context.Background()

	scenarioDir, err := filepath.Abs(
		filepath.Join(
			"..",
			"..",
			"..",
			"scenarios",
			scenario,
		),
	)
	if err != nil {
		return nil, err
	}

	// Generates:
	//   server.ks
	//   client.ks
	//   test-descriptor-bundle.zip
	if err := PrepareScenario(scenarioDir); err != nil {
		return nil, err
	}

	req := testcontainers.ContainerRequest{

		Image: KublingImage,

		ExposedPorts: []string{
			"8282/tcp",
			"50051/tcp",
		},

		Env: map[string]string{

			"DESCRIPTOR_BUNDLE": "/kbl/" + BundleName,

			"APP_CONFIG": "/kbl/app-config.yaml",

			"SCRIPT_LOG_LEVEL": "DEBUG",

			"INTERNAL_API_KEY": "integration-tests",

			"GRPC_SSL_CERT": "/kbl/" + ServerKeyStore,

			"GRPC_SSL_PASS": ServerPassword,
		},

		HostConfigModifier: func(
			host *containertypes.HostConfig,
		) {

			host.Binds = append(
				host.Binds,
				fmt.Sprintf(
					"%s:/kbl",
					scenarioDir,
				),
			)

		},

		WaitingFor: wait.ForAll(
			wait.ForListeningPort("50051/tcp"),
			wait.ForHTTP("/observe/health").
				WithPort("8282/tcp").
				WithStatusCodeMatcher(func(status int) bool {
					return status == 200
				}),
		),
	}

	c, err :=
		testcontainers.GenericContainer(
			ctx,
			testcontainers.GenericContainerRequest{
				ContainerRequest: req,
				Started:          true,
			},
		)

	if err != nil {
		return nil, err
	}

	c.FollowOutput(&StdoutLogConsumer{})

	err = c.StartLogProducer(context.Background())
	if err != nil {
		return nil, err
	}

	host, err := c.Host(ctx)

	if err != nil {
		_ = c.Terminate(ctx)
		return nil, err
	}

	port, err := c.MappedPort(
		ctx,
		"50051",
	)

	if err != nil {
		_ = c.Terminate(ctx)
		return nil, err
	}

	return &Runtime{

		container: c,

		Address: fmt.Sprintf(
			"%s:%s",
			host,
			port.Port(),
		),
	}, nil

}

func (r *Runtime) Close() error {

	return r.container.Terminate(
		context.Background(),
	)

}

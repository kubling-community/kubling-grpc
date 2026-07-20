package contracts

import (
	"os"
	"testing"

	"github.com/kubling-community/kubling-grpc/sdk-go/tests/support"
)

var runtime *support.Runtime

func TestMain(
	m *testing.M,
) {

	var err error

	runtime, err = support.StartRuntime(
		support.MinimalScenario,
	)
	if err != nil {
		panic(err)
	}

	code := m.Run()

	_ = runtime.Close()

	os.Exit(code)

}

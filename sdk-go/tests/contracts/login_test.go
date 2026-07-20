package contracts

import (
	"testing"
)

func TestLogin(
	t *testing.T,
) {

	client :=
		runtime.Client(
			t,
		)

	if err := client.Logout(); err != nil {
		t.Fatal(err)
	}

}

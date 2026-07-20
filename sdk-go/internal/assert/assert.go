package assert

import (
	"fmt"
	"strings"
)

func AssertTrue(
	value bool,
	message string,
) error {

	if !value {
		return fmt.Errorf(
			"assertion failed: %s",
			message,
		)
	}

	return nil
}

func AssertEquals(
	expected interface{},
	actual interface{},
) error {

	if expected != actual {

		return fmt.Errorf(
			"expected %v but got %v",
			expected,
			actual,
		)

	}

	return nil

}

func Fail(
	message string,
) error {

	return fmt.Errorf(
		"failure: %s",
		message,
	)

}

func AssertErrorContains(
	err error,
	expected string,
) error {

	if err == nil {
		return fmt.Errorf(
			"expected error containing '%s'",
			expected,
		)
	}

	if !strings.Contains(
		err.Error(),
		expected,
	) {

		return fmt.Errorf(
			"expected '%s' but got '%s'",
			expected,
			err,
		)

	}

	return nil

}

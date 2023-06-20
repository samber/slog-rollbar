package slogrollbar

import (
	"testing"

	// "go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	// commented because the rollbar library is leaking a coroutine
	// goleak.VerifyTestMain(m)
}

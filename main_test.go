package slogrollbar

import (
	"testing"

	"go.uber.org/goleak"
)

var knownGoroutineLeaks = []goleak.Option{
	goleak.IgnoreTopFunction("github.com/rollbar/rollbar-go.NewAsyncTransport.func1"),
}

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m, knownGoroutineLeaks...)
}

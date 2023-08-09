package slogrollbar

import (
	"runtime"
	"strings"
)

// Stolen from https://github.com/heroku/rollrus/blob/master/hook.go

// framesToSkip returns the number of caller frames to skip
// to get a stack trace that excludes rollrus and logrus.
func framesToSkip(rollrusSkip int) int {
	// skip 1 to get out of this function
	skip := rollrusSkip + 1

	// to get out of logrus, the amount can vary
	// depending on how the user calls the log functions
	// figure it out dynamically by skipping until
	// we're out of the logrus package
	for i := skip; ; i++ {
		_, file, _, ok := runtime.Caller(i)
		if !ok || (!strings.Contains(file, "slog") && !strings.Contains(file, "log/slog")) {
			skip = i
			break
		}
	}

	// rollbar-go is skipping too few frames (2)
	// subtract 1 since we're currently working from a function
	return skip + 2 - 1
}

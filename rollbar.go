package slogrollbar

import (
	"log/slog"

	"github.com/rollbar/rollbar-go"
)

var LogLevels = map[slog.Level]string{
	slog.LevelDebug: rollbar.DEBUG,
	slog.LevelInfo:  rollbar.INFO,
	slog.LevelWarn:  rollbar.WARN,
	slog.LevelError: rollbar.ERR,
}

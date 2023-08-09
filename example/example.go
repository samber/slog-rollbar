package main

import (
	"fmt"
	"time"

	"github.com/rollbar/rollbar-go"
	slogrollbar "github.com/samber/slog-rollbar"

	"log/slog"
)

func main() {
	token := "xxxxx"
	env := "production"
	version := "v1"
	host := "127.0.0.1"
	project := "samber/slog-rollbar/example"

	client := rollbar.NewAsync(token, env, version, host, project)
	defer client.Close()

	logger := slog.New(slogrollbar.Option{Level: slog.LevelDebug, Client: client}.NewRollbarHandler())

	logger.
		With(
			slog.Group("user",
				slog.String("id", "user-123"),
				slog.Time("created_at", time.Now()),
			),
		).
		With("error", fmt.Errorf("an error")).
		Error("a message")
}

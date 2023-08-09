package slogrollbar

import (
	"context"

	"github.com/rollbar/rollbar-go"

	"log/slog"
)

type Option struct {
	// log level (default: debug)
	Level slog.Leveler

	// Rollbar client
	Client *rollbar.Client

	// optional: customize Rollbar event builder
	Converter Converter
}

func (o Option) NewRollbarHandler() slog.Handler {
	if o.Level == nil {
		o.Level = slog.LevelDebug
	}

	return &RollbarHandler{
		option: o,
		attrs:  []slog.Attr{},
		groups: []string{},
	}
}

type RollbarHandler struct {
	option Option
	attrs  []slog.Attr
	groups []string
}

func (h *RollbarHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.option.Level.Level()
}

func (h *RollbarHandler) Handle(ctx context.Context, record slog.Record) error {
	converter := DefaultConverter
	if h.option.Converter != nil {
		converter = h.option.Converter
	}

	extra, err := converter(h.attrs, record)

	switch record.Level {
	case slog.LevelDebug:
		h.option.Client.MessageWithExtras(rollbar.DEBUG, record.Message, extra)
	case slog.LevelInfo:
		h.option.Client.MessageWithExtras(rollbar.INFO, record.Message, extra)
	case slog.LevelWarn:
		h.option.Client.MessageWithExtras(rollbar.WARN, record.Message, extra)
	case slog.LevelError:
		if err != nil {
			skip := framesToSkip(2)
			h.option.Client.ErrorWithStackSkipWithExtras(rollbar.ERR, err, skip, extra)
		} else {
			h.option.Client.MessageWithExtras(rollbar.ERR, record.Message, extra)
		}
	}

	return nil
}

func (h *RollbarHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &RollbarHandler{
		option: h.option,
		attrs:  appendAttrsToGroup(h.groups, h.attrs, attrs),
		groups: h.groups,
	}
}

func (h *RollbarHandler) WithGroup(name string) slog.Handler {
	return &RollbarHandler{
		option: h.option,
		attrs:  h.attrs,
		groups: append(h.groups, name),
	}
}

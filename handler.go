package slogrollbar

import (
	"context"

	"github.com/rollbar/rollbar-go"
	slogcommon "github.com/samber/slog-common"

	"log/slog"
)

type Option struct {
	// log level (default: debug)
	Level slog.Leveler

	// Rollbar client
	Client *rollbar.Client

	// optional: customize Rollbar event builder
	Converter Converter

	// optional: see slog.HandlerOptions
	AddSource   bool
	ReplaceAttr func(groups []string, a slog.Attr) slog.Attr
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

var _ slog.Handler = (*RollbarHandler)(nil)

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

	extra, err := converter(h.option.AddSource, h.option.ReplaceAttr, h.attrs, h.groups, &record)

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
		attrs:  slogcommon.AppendAttrsToGroup(h.groups, h.attrs, attrs...),
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

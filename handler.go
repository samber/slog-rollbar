package slogrollbar

import (
	"context"
	"time"

	"github.com/rollbar/rollbar-go"
	slogcommon "github.com/samber/slog-common"

	"log/slog"
)

type Option struct {
	// log level (default: debug)
	Level slog.Leveler

	// Rollbar client
	Client  *rollbar.Client
	Timeout time.Duration // default: 10s

	// optional: customize Rollbar event builder
	Converter Converter
	// optional: fetch attributes from context
	AttrFromContext []func(ctx context.Context) []slog.Attr

	// optional: see slog.HandlerOptions
	AddSource   bool
	ReplaceAttr func(groups []string, a slog.Attr) slog.Attr
}

func (o Option) NewRollbarHandler() slog.Handler {
	if o.Level == nil {
		o.Level = slog.LevelDebug
	}

	if o.Timeout == 0 {
		o.Timeout = 10 * time.Second
	}

	if o.Converter == nil {
		o.Converter = DefaultConverter
	}

	if o.AttrFromContext == nil {
		o.AttrFromContext = []func(ctx context.Context) []slog.Attr{}
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
	fromContext := slogcommon.ContextExtractor(ctx, h.option.AttrFromContext)
	extra := h.option.Converter(h.option.AddSource, h.option.ReplaceAttr, append(h.attrs, fromContext...), h.groups, &record)
	level := LogLevels[record.Level]

	ctx, cancel := context.WithTimeout(context.Background(), h.option.Timeout)
	defer cancel()

	// if level == rollbar.ERR || level == rollbar.CRIT {
	// 	skip := framesToSkip(2)
	// 	h.option.Client.ErrorWithStackSkipWithExtrasAndContext(ctx, rollbar.ERR, err, skip, extra)
	//  return nil
	// }

	h.option.Client.MessageWithExtrasAndContext(ctx, level, record.Message, extra)

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
	// https://cs.opensource.google/go/x/exp/+/46b07846:slog/handler.go;l=247
	if name == "" {
		return h
	}

	return &RollbarHandler{
		option: h.option,
		attrs:  h.attrs,
		groups: append(h.groups, name),
	}
}

package slogrollbar

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/rollbar/rollbar-go"
	slogcommon "github.com/samber/slog-common"
)

type wrappedError struct {
	msg string
	error
}

func (w wrappedError) Unwrap() error {
	return w.error
}

func (w wrappedError) Error() string {
	return w.msg
}

type Option struct {
	// log level (default: debug)
	Level slog.Leveler

	// Rollbar client
	Client     *rollbar.Client
	Timeout    time.Duration // default: 10s
	SkipFrames *int          // default: 2

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

	if o.ReplaceAttr == nil {
		o.ReplaceAttr = defaultReplaceAttr
	}

	if o.SkipFrames == nil {
		o.SkipFrames = new(int)
		*o.SkipFrames = 2
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
	extra := h.option.Converter(h.option.AddSource, h.option.ReplaceAttr,
		append(h.attrs, fromContext...), h.groups, &record)
	level := LogLevels[record.Level]

	ctx, cancel := context.WithTimeout(context.Background(), h.option.Timeout)
	defer cancel()

	// extract error and request from slog.Record
	var r *http.Request
	var err error

	record.Attrs(func(a slog.Attr) bool {
		if err != nil && r != nil {
			return false
		}

		switch v := a.Value.Any().(type) {
		case *http.Request:
			// Pass the request to rollbar for additinoal context
			r = v
		case error:
			// Keep the original message, but report as an error to Rollbar
			// This enables stack tracing if the error has it.
			err = wrappedError{msg: record.Message, error: v}
		}
		return true
	})

	if err != nil {
		if r == nil {
			h.option.Client.ErrorWithStackSkipWithExtrasAndContext(ctx, level, err, *h.option.SkipFrames, extra)
		} else {
			h.option.Client.RequestErrorWithStackSkipWithExtrasAndContext(ctx, level, r, err, *h.option.SkipFrames, extra)
		}
	} else {
		if r == nil {
			h.option.Client.MessageWithExtrasAndContext(ctx, level, record.Message, extra)
		} else {
			h.option.Client.RequestMessageWithExtrasAndContext(ctx, level, r, record.Message, extra)
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

func defaultReplaceAttr(groups []string, a slog.Attr) slog.Attr {
	// Leave group untouched
	if len(groups) > 1 {
		return a
	}

	// rollbar does not know how send http.Request objects
	if _, ok := a.Value.Any().(*http.Request); ok {
		return slog.Attr{}
	}

	return a
}

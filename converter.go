package slogrollbar

import (
	"golang.org/x/exp/slog"
)

type Converter func(loggerAttr []slog.Attr, record slog.Record) (map[string]any, error)

func DefaultConverter(loggerAttr []slog.Attr, record slog.Record) (map[string]any, error) {
	output := attrsToValue(loggerAttr)

	record.Attrs(func(attr slog.Attr) bool {
		output[attr.Key] = attrToValue(attr)
		return true
	})

	if v, ok := output["error"]; ok {
		if err, ok := v.(error); ok {
			delete(output, "error")
			return output, err
		}
	}

	return output, nil
}

func attrsToValue(attrs []slog.Attr) map[string]any {
	output := map[string]any{}
	for i := range attrs {
		output[attrs[i].Key] = attrToValue(attrs[i])
	}
	return output
}

func attrToValue(attr slog.Attr) any {
	v := attr.Value
	kind := attr.Value.Kind()

	switch kind {
	case slog.KindAny:
		return v.Any()
	case slog.KindLogValuer:
		return v.LogValuer().LogValue().Any()
	case slog.KindGroup:
		return attrsToValue(v.Group())
	case slog.KindInt64:
		return v.Int64()
	case slog.KindUint64:
		return v.Uint64()
	case slog.KindFloat64:
		return v.Float64()
	case slog.KindString:
		return v.String()
	case slog.KindBool:
		return v.Bool()
	case slog.KindDuration:
		return v.Duration()
	case slog.KindTime:
		return v.Time().UTC()
	default:
		return v.Any()
	}
}

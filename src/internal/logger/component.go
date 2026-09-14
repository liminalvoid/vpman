package logger

import (
	"context"
	"log/slog"
	"fmt"
)

type componentHandler struct {
	slog.Handler
	component string
	grouped   bool
	color     bool
}

func (h *componentHandler) Handle(ctx context.Context, r slog.Record) error {
	if h.component != "" {
		prefix := fmt.Sprintf("[%s]", h.component)
		if h.color {
			prefix = fmt.Sprintf("\033[36m%s\033[0m", prefix)
		}

		r.Message = fmt.Sprintf("%s %s", prefix, r.Message)
	}

	return h.Handler.Handle(ctx, r)
}

func (h *componentHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h2 := *h
	rest := make([]slog.Attr, 0, len(attrs))
	for _, a := range attrs {
		if !h.grouped && a.Key == "component" {
			h2.component = a.Value.String()

			continue
		}

		rest = append(rest, a)
	}

	h2.Handler = h.Handler.WithAttrs(rest)

	return &h2
}

func (h *componentHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	h2 := *h
	h2.Handler = h.Handler.WithGroup(name)
	h2.grouped = true

	return &h2
}

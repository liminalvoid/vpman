package logger

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/lmittmann/tint"
)

func New(level, format string) *slog.Logger {
	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(level)); err != nil {
		lvl = slog.LevelInfo
	}
	opts := &slog.HandlerOptions{
		Level: lvl,
		AddSource: lvl == slog.LevelDebug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if len(groups) > 0 {
				return a
			}

			switch a.Key {
			case slog.TimeKey:
				return slog.String(a.Key, a.Value.Time().Format("15:04:05.000"))
			case slog.SourceKey:
				src := a.Value.Any().(*slog.Source)

				return slog.String(
					a.Key,
					fmt.Sprintf("%s:%d", filepath.Base(src.File), src.Line),
				)
			}

			return a
		},
	}

	var h slog.Handler
	switch format {
	case "json":
		h = slog.NewJSONHandler(os.Stdout, opts)
	case "pretty":
		h = &componentHandler{
			Handler: tint.NewTextHandler(os.Stdout, &tint.Options{
				Level: lvl,
				TimeFormat: time.Kitchen,
				ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
					return a
				},
			}),
			color: true,
		}
	default:
		h = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(h)
}

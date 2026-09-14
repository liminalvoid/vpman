package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"path/filepath"
)

func formatLogEntry(groups []string, a slog.Attr) slog.Attr {
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
}

func New(level, format string) *slog.Logger {
	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(level)); err != nil {
		lvl = slog.LevelInfo
	}
	opts := &slog.HandlerOptions{
		Level: lvl,
		AddSource: lvl == slog.LevelDebug,
		ReplaceAttr: formatLogEntry,
	}

	var h slog.Handler
	if format == "json" {
		h = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		h = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(h)
}

// Telego adapts *slog.Logger to telego.Logger and redacts the bot token
type Telego struct {
	log *slog.Logger
	redactor *strings.Replacer
}

func NewTelego(l *slog.Logger, token string) *Telego {
	return &Telego{
		log: l.With("component", "telego"),
		redactor: strings.NewReplacer(token, "BOT_TOKEN"),
	}
}

func (t *Telego) Debugf(format string, args ...any) {
	if !t.log.Enabled(context.Background(), slog.LevelDebug) {
		return
	}
	t.log.Debug(t.redactor.Replace(fmt.Sprintf(format, args...)))
}

func (t *Telego) Errorf(format string, args ...any) {
	t.log.Error(t.redactor.Replace(fmt.Sprintf(format, args...)))
}

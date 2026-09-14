package logger

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
)

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

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"manager/internal/configgen"
	"manager/internal/executor"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// TODO: Initialize modules (DB, Telegram bot, config manager)
	
}

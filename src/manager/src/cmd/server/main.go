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

	cli, err := executor.New("5e13c8f1326929ea701a5a5fe41a4b7aab4ceb37cc78629d13fcb2f2e36d71fd", "127.0.0.1")
	if err != nil {
		fmt.Printf("failed to create container executor: %s", err)

		return
	}

	awg := &configgen.Awg{}
	cg := configgen.New(awg)

	cg.Create(ctx, cli)
}

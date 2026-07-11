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

	executorConfig := &executor.Config{
		User:        "root",
		HostAddr:    "192.168.0.1",
		HostPort:    "22",
		ContainerID: "5e13c8f1326929ea701a5a5fe41a4b7aab4ceb37cc78629d13fcb2f2e36d71fd",
	}
	cli, err := executor.New(executorConfig)
	if err != nil {
		fmt.Printf("failed to create container executor: %s", err)

		return
	}

	wg := &configgen.Awg{}
	cg := configgen.New(wg)

	cg.Create(ctx, cli)
}

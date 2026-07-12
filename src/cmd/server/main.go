package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/joho/godotenv"

	"vpnmanager/internal/configgen"
	"vpnmanager/internal/executor"
	"vpnmanager/internal/tgbot"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Load dotenv
	err := godotenv.Load(".env.development")
	if err != nil {
		fmt.Println("Error loading .env file")
		os.Exit(1)
	}

	awg := &configgen.Awg{}
	cg := configgen.New(awg)

	ceCfg := executor.Config{
		User:        "root",
		HostAddr:    "192.168.0.127",
		HostPort:    "22",
		ContainerID: "dfksjwewfef11234",
	}
	ce, err := executor.New(&ceCfg)
	if err != nil {
		fmt.Printf("failed to start executor: %s\n", err)
		os.Exit(1)
	}

	botToken := os.Getenv("TG_TOKEN")
	idSizeBytes := 64
	adminId, err := strconv.ParseInt(os.Getenv("ME_ID"), 10, idSizeBytes)
	if err != nil {
		fmt.Println("The administrator ID wasn't parsed")
		os.Exit(1)
	}
	bhCfg := tgbot.Config{
		Token:   botToken,
		AdminID: adminId,
	}
	bh, err := tgbot.New(ctx, bhCfg, ce, cg)
	if err != nil {
		fmt.Printf("failed to create Telegram bot: %s\n", err)
		os.Exit(1)
	}

	defer func() { bh.Stop() }()
	err = bh.Start()
	if err != nil {
		fmt.Printf("failed to start Telegram bot: %s\n", err)
		os.Exit(1)
	}
}

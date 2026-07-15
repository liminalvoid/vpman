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
	"vpnmanager/internal/db"
	"vpnmanager/internal/executor"
	"vpnmanager/internal/repository"
	"vpnmanager/internal/service"
	"vpnmanager/internal/tgbot"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Load dotenv
	err := godotenv.Load(".env.dev")
	if err != nil {
		fmt.Println("Error loading .env file")
		os.Exit(1)
	}

	// Initialize container executor
	ceCfg := executor.Config{
		User:        "root",
		HostAddr:    "127.0.0.1",
		HostPort:    "22",
		ContainerID: "5e13c8f1326929ea701a5a5fe41a4b7aab4ceb37cc78629d13fcb2f2e36d71fd",
		FromEnv:     true,
	}
	ce, err := executor.New(&ceCfg)
	if err != nil {
		fmt.Printf("failed to start executor: %s\n", err)
		os.Exit(1)
	}

	// Initialize config generator
	wg := &configgen.Wg{}
	cg := configgen.New(wg, ce)

	// Initialize DB module
	dbCfg := db.Config{
		Username:     "postgres",
		Password:     "postgres",
		HostAddr:     "localhost",
		HostPort:     5432,
		DatabaseName: "vpn",
	}
	db, err := db.New(ctx, &dbCfg)
	if err != nil {
		fmt.Printf("failed to init DB: %s\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Initialize repos
	configRepo := repository.NewConfig(db.Pool())

	// Initialize peer service
	ps := service.NewPeer(cg)

	// Initialize Telegram bot
	botToken := os.Getenv("TG_TOKEN")
	idSizeBytes := 64
	adminId, err := strconv.ParseInt(os.Getenv("ME_ID"), 10, idSizeBytes)
	if err != nil {
		fmt.Println("The administrator ID wasn't parsed")
		os.Exit(1)
	}
	bCfg := tgbot.Config{
		Token:   botToken,
		AdminID: adminId,
	}
	b, err := tgbot.New(ctx, bCfg, ps)
	if err != nil {
		fmt.Printf("failed to create Telegram bot: %s\n", err)
		os.Exit(1)
	}
	defer func() { b.Stop() }()

	// Start bot polling
	err = b.Start()
	if err != nil {
		fmt.Printf("failed to start Telegram bot: %s\n", err)
		os.Exit(1)
	}
}

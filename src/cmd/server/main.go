package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"vpnmanager/internal/db"
	"vpnmanager/internal/executor"
	"vpnmanager/internal/repository"
	"vpnmanager/internal/scheduler"
	"vpnmanager/internal/service"
	"vpnmanager/internal/tgbot"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	err := godotenv.Load(".env.dev")
	if err != nil {
		fmt.Println("Error loading .env file")
		os.Exit(1)
	}

	cr := executor.NewRegistry(executor.NewExecutor)
	defer cr.Close()

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
	protocolRepo := repository.NewProtocol(db.Pool())
	serverRepo := repository.NewServer(db.Pool())
	userRepo := repository.NewUser(db.Pool())
	subscriptionRepo := repository.NewSubscription(db.Pool())
	accessCodeRepo := repository.NewAccessCode(db.Pool())

	// Initialize services
	peerService := service.NewConfig(
		configRepo,
		protocolRepo,
		userRepo,
		subscriptionRepo,
		cr,
	)
	serverService := service.NewServer(serverRepo, cr)
	subscriptionService := service.NewSubscription(
		subscriptionRepo,
		userRepo,
		configRepo,
		serverRepo,
		cr,
	)
	userService := service.NewUser(userRepo)
	accessCodeService := service.NewAccessCodeService(accessCodeRepo)

	s := scheduler.New(
		subscriptionService,
		5 * time.Minute,
		100,
		30 * 24 * time.Hour,
	)
	go s.Run(ctx)

	botToken := os.Getenv("TG_TOKEN")
	bCfg := tgbot.Config{
		Token:   botToken,
	}
	b, err := tgbot.New(
		ctx,
		bCfg,
		peerService,
		serverService,
		userService,
		accessCodeService,
	)
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

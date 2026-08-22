package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/mymmrac/telego"

	"vpnmanager/internal/db"
	"vpnmanager/internal/executor"
	"vpnmanager/internal/repository"
	"vpnmanager/internal/scheduler"
	"vpnmanager/internal/service"
	"vpnmanager/internal/tgbot"
	"vpnmanager/internal/tgbot/newsletter"
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

	// Initialize Telegram bot
	// FIXME: Remove this and think how to handle logs in prod
	botToken := os.Getenv("TG_TOKEN")
	tg, err := telego.NewBot(botToken, telego.WithDefaultDebugLogger())

	nl := newsletter.New(tg)

	// Initialize repos
	configRepo := repository.NewConfig(db.Pool())
	protocolRepo := repository.NewProtocol(db.Pool())
	serverRepo := repository.NewServer(db.Pool())
	userRepo := repository.NewUser(db.Pool())
	subscriptionRepo := repository.NewSubscription(db.Pool())
	accessCodeRepo := repository.NewAccessCode(db.Pool())

	// Initialize services
	configService := service.NewConfig(
		configRepo,
		protocolRepo,
		userRepo,
		subscriptionRepo,
		serverRepo,
		cr,
	)
	serverService := service.NewServer(serverRepo, cr)
	userService := service.NewUser(userRepo, subscriptionRepo)
	accessCodeService := service.NewAccessCodeService(accessCodeRepo)
	subscriptionService := service.NewSubscription(
		cr,
		nl,
		subscriptionRepo,
		userRepo,
		configRepo,
		serverRepo,
	)

	b, err := tgbot.New(
		ctx,
		tg,
		configService,
		serverService,
		userService,
		accessCodeService,
		subscriptionService,
	)
	if err != nil {
		fmt.Printf("failed to create Telegram bot: %s\n", err)
		os.Exit(1)
	}
	defer func() { b.Stop() }()

	s := scheduler.New(
		subscriptionService,
		configService,
		5 * time.Minute,
		100,
		30 * 24 * time.Hour,
	)
	go s.Run(ctx)

	// Start bot polling
	err = b.Start()
	if err != nil {
		fmt.Printf("failed to start Telegram bot: %s\n", err)
		os.Exit(1)
	}
}

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/mymmrac/telego"

	"vpnmanager/internal/config"
	"vpnmanager/internal/db"
	"vpnmanager/internal/executor"
	"vpnmanager/internal/health"
	"vpnmanager/internal/repository"
	"vpnmanager/internal/scheduler"
	"vpnmanager/internal/service"
	"vpnmanager/internal/tgbot"
	"vpnmanager/internal/tgbot/newsletter"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		os.Exit(runHealthcheck())
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}

	cr := executor.NewRegistry(executor.NewExecutor)
	defer cr.Close()

	db, err := db.New(ctx, cfg.DatabaseURL)
	if err != nil {
		fmt.Printf("failed to init DB: %s\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Migrate(ctx); err != nil {
		fmt.Printf("failed to migrate DB: %s\n", err)
		os.Exit(1)
	}

	// Initialize Telegram bot
	// FIXME: Remove this and think how to handle logs in prod
	tg, err := telego.NewBot(cfg.TelegramToken, telego.WithDefaultDebugLogger())
	if err != nil {
		fmt.Printf("failed to create Telegram client: %s\n", err)
		os.Exit(1)
	}

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
		5*time.Minute,
		100,
		30*24*time.Hour,
	)
	go s.Run(ctx)

	h := health.New(cfg.HealthAddr, db.Pool())
	go h.Run(ctx)

	// Start bot polling. Returns when ctx is cancelled (SIGTERM from Swarm),
	// which is what keeps the long-polling singleton from overlapping with its
	// replacement during a stop-first rolling update.
	err = b.Start()
	if err != nil {
		fmt.Printf("failed to start Telegram bot: %s\n", err)
		os.Exit(1)
	}
}

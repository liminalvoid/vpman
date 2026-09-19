package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mymmrac/telego"

	"vpnmanager/internal/config"
	"vpnmanager/internal/db"
	"vpnmanager/internal/executor"
	"vpnmanager/internal/health"
	"vpnmanager/internal/logger"
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

	log := logger.New(cfg.LogLevel, cfg.LogFormat)
	slog.SetDefault(log)

	// Initialize Telegram bot
	tg, err := telego.NewBot(
		cfg.TelegramToken,
		telego.WithLogger(logger.NewTelego(log, cfg.TelegramToken)),
	)
	if err != nil {
		fmt.Printf("failed to create Telegram client: %s\n", err)
		os.Exit(1)
	}

	// Initialize newsletter service for sending notifications
	nl := newsletter.New(log.With("component", "newsletter"), tg)

	// Initialize repos
	repos := repository.New(db.Pool())

	// Initialize services
	services := service.New(log, repos, cr, nl)

	b, err := tgbot.New(
		ctx,
		tg,
		services.Config,
		services.Server,
		services.User,
		services.AccessCode,
		services.Subscription,
	)
	if err != nil {
		fmt.Printf("failed to create Telegram bot: %s\n", err)
		os.Exit(1)
	}
	defer func() { b.Stop() }()

	scheduled := map[string]scheduler.Scheduled{
		"past_due":     services.Subscription.SetPastDue,
		"remove_stale": services.Config.RemoveStale,
	}
	s := scheduler.New(
		log.With("component", "scheduler"),
		scheduled,
		services.Subscription,
		services.Config,
		5*time.Minute,
		100,
		30*24*time.Hour,
	)
	go s.Run(ctx)

	h := health.New(log.With("component", "health"), cfg.HealthAddr, db.Pool())
	go h.Run(ctx)

	err = b.Start()
	if err != nil {
		fmt.Printf("failed to start Telegram bot: %s\n", err)
		os.Exit(1)
	}
}

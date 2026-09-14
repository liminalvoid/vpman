package service

import (
	"log/slog"

	"vpnmanager/internal/executor"
	"vpnmanager/internal/repository"
	"vpnmanager/internal/tgbot/newsletter"
)

type Services struct {
	AccessCode   *AccessCodeService
	Config       *ConfigService
	Server       *ServerService
	Subscription *SubscriptionService
	User         *UserService
}

func New(
	log *slog.Logger,
	repos *repository.Repositories,
	cr *executor.Registry,
	nl *newsletter.Newsletter,
) *Services {
	return &Services{
		AccessCode: NewAccessCode(log.With("component", "service.access_code"), repos.AccessCode),
		Config: NewConfig(
			log.With("component", "service.config"),
			repos.Config,
			repos.Protocol,
			repos.User,
			repos.Subscription,
			repos.Server,
			cr,
		),
		Server: NewServer(log.With("component", "service.server"), repos.Server, cr),
		Subscription: NewSubscription(
			log.With("component", "service.subscription"),
			cr,
			nl,
			repos.Subscription,
			repos.User,
			repos.Config,
			repos.Server,
		),
		User: NewUser(log.With("component", "service.user"), repos.User, repos.Subscription),
	}
}

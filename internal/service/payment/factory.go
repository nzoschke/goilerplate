package payment

import (
	"fmt"
	"log/slog"

	"github.com/templui/goilerplate/internal/config"
	"github.com/templui/goilerplate/internal/model"
	"github.com/templui/goilerplate/internal/service"
)

// NewProvider creates a payment provider based on configuration
func NewProvider(cfg *config.Config, subscriptionService *service.SubscriptionService) (Provider, error) {
	provider := cfg.PaymentProvider

	slog.Info("initializing payment provider", "provider", provider)

	switch provider {
	case model.ProviderPolar:
		if cfg.PolarAPIKey == "" {
			return nil, fmt.Errorf("POLAR_API_KEY is required when using Polar provider")
		}
		return NewPolarProvider(cfg, subscriptionService), nil

	case model.ProviderStripe:
		if cfg.StripeSecretKey == "" {
			return nil, fmt.Errorf("STRIPE_SECRET_KEY is required when using Stripe provider")
		}
		if cfg.StripeWebhookSecret == "" {
			return nil, fmt.Errorf("STRIPE_WEBHOOK_SECRET is required when using Stripe provider")
		}
		return NewStripeProvider(cfg, subscriptionService), nil

	default:
		return nil, fmt.Errorf("unknown payment provider: %s (supported: polar, stripe)", provider)
	}
}

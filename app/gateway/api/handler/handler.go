package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/cep21/circuit/v3"
	"github.com/cep21/circuit/v3/closers/hystrix"
	"github.com/go-chi/chi/v5"

	"github.com/chatbot-go/app/config"
	"github.com/chatbot-go/app/domain/usecase"
	"github.com/chatbot-go/app/gateway/client/twilio"
)

type Handler struct {
	circuitManager *circuit.Manager
	cfg            config.Config
	useCase        useCase
	cache          cache
}

func New(cfg config.Config, useCase useCase, cache cache) Handler {
	hystrixFactory := hystrix.Factory{
		ConfigureOpener: hystrix.ConfigureOpener{
			ErrorThresholdPercentage: int64(cfg.CircuitBreaker.ErrorPercentThreshold),
			RequestVolumeThreshold:   int64(cfg.CircuitBreaker.RequestVolumeThreshold),
		},
		ConfigureCloser: hystrix.ConfigureCloser{
			SleepWindow: cfg.CircuitBreaker.SleepWindow,
		},
	}

	defaultFactory := func(_ string) circuit.Config {
		return circuit.Config{
			Execution: circuit.ExecutionConfig{
				MaxConcurrentRequests: int64(cfg.CircuitBreaker.MaxConcurrentRequests),
				Timeout:               cfg.CircuitBreaker.Timeout,
			},
		}
	}

	circuitManager := &circuit.Manager{
		DefaultCircuitProperties: []circuit.CommandPropertiesConstructor{
			defaultFactory,
			hystrixFactory.Configure,
		},
	}

	return Handler{
		circuitManager: circuitManager,
		cfg:            cfg,
		useCase:        useCase,
		cache:          cache,
	}
}

func RegisterHealthCheckRoute(router chi.Router) {
	router.Get("/healthcheck", func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusOK)
	})
}

func RegisterPublicRoutes(
	router chi.Router,
	cfg config.Config,
	useCase useCase,
	cache cache,
	twilioClient *twilio.Client,
) {
	handler := New(cfg, useCase, cache)

	handler.WebhooksTwilioSetup(router, twilioClient)
}

type cache interface {
	Exists(ctx context.Context, key string) (bool, error)
	Get(ctx context.Context, key string, objByRef any) error
	Set(ctx context.Context, key string, obj any, ttl time.Duration) error
}

type useCase interface {
	EnqueueTwilioWebhook(ctx context.Context, input usecase.EnqueueTwilioWebhookInput) error
}

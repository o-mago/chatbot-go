package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/chatbot-go/app/config"
	"github.com/chatbot-go/app/domain/usecase"
	"github.com/chatbot-go/app/gateway/api/handler"
	"github.com/chatbot-go/app/gateway/api/middleware"
	"github.com/chatbot-go/app/gateway/client/twilio"
	"github.com/chatbot-go/app/gateway/redis"
)

type API struct {
	Handler      http.Handler
	cfg          config.Config
	useCase      *usecase.UseCase
	redisClient  *redis.Client
	twilioClient *twilio.Client
}

func BasicHandler() http.Handler {
	router := chi.NewMux()
	handler.RegisterHealthCheckRoute(router)

	return router
}

func New(cfg config.Config, redisClient *redis.Client, useCase *usecase.UseCase, twilioClient *twilio.Client) *API {
	api := &API{
		cfg:          cfg,
		useCase:      useCase,
		redisClient:  redisClient,
		twilioClient: twilioClient,
	}

	api.setupRouter()

	return api
}

func (api *API) setupRouter() {
	router := chi.NewRouter()

	if api.cfg.Development {
		router.Use(middleware.Logger)
	}

	router.Use(
		middleware.CORS,
		middleware.CleanPath,
		middleware.StripSlashes,
		middleware.HeadersToContext,
		middleware.Recoverer,
	)

	api.registerRoutes(router)

	api.Handler = router
}

func (api *API) registerRoutes(router *chi.Mux) {
	handler.RegisterHealthCheckRoute(router)

	router.Route("/api/v1/chatbot", func(publicRouter chi.Router) {
		handler.RegisterPublicRoutes(
			publicRouter,
			api.cfg,
			api.useCase,
			api.redisClient,
			api.twilioClient,
		)
	})
}

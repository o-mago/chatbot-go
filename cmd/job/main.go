package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"

	"github.com/chatbot-go/app"
	"github.com/chatbot-go/app/config"
	"github.com/chatbot-go/app/gateway/cronjob"
	"github.com/chatbot-go/app/gateway/postgres"
	"github.com/chatbot-go/app/gateway/redis"
	"github.com/chatbot-go/app/gateway/sqs"
	"github.com/chatbot-go/app/telemetry"
)

// Injected on build via ldflags.
var (
	BuildTime   = "undefined"
	BuildCommit = "undefined"
	BuildTag    = "undefined"
)

func main() {
	mainCtx := context.Background()

	// Config
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("failed to load configurations: %v", err)
	}

	// Logger
	telemetry.SetLogger(cfg.Development,
		slog.String("build_time", BuildTime),
		slog.String("build_commit", BuildCommit),
		slog.String("build_tag", BuildTag),
	)

	// Open Telemetry
	otel, err := telemetry.NewOtel(mainCtx, cfg.Otel, string(cfg.Environment), BuildTag)
	if err != nil {
		log.Fatalf("failed to start otel: %v", err)
	}

	ctx := telemetry.ContextWithTracer(mainCtx, otel.Tracer)

	// Postgres
	postgresClient, err := postgres.New(context.Background(), cfg.Postgres)
	if err != nil {
		log.Fatalf("failed to start postgres: %v", err)
	}

	// Redis
	redisClient, err := redis.New(ctx, cfg.Redis)
	if err != nil {
		log.Fatalf("failed to start redis: %v", err)
	}

	// SQS
	_, sqsEnqueuer, err := sqs.New(ctx, cfg.SQS, cfg.Development)
	if err != nil {
		log.Fatalf("failed to start sqs: %v", err)
	}

	// Application
	appl, err := app.New(cfg, postgresClient, redisClient, sqsEnqueuer)
	if err != nil {
		log.Fatalf("failed to start application: %v", err)
	}

	// Cronjob
	cronjob := cronjob.New(appl.UseCase)

	// Graceful Shutdown
	stopCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	group, groupCtx := errgroup.WithContext(stopCtx)

	//nolint:wrapcheck
	group.Go(func() error {
		log.Printf("starting job cronjob")

		if err := cronjob.RunContext(ctx, os.Args); err != nil {
			return err
		}

		stop()

		return nil
	})
	//nolint:contextcheck
	group.Go(func() error {
		<-groupCtx.Done()

		log.Printf("stopping job; interrupt signal received")

		timeoutCtx, cancel := context.WithTimeout(context.Background(), cfg.App.GracefulShutdownTimeout)
		defer cancel()

		var errs error

		if err := otel.Close(timeoutCtx); err != nil {
			errs = errors.Join(errs, fmt.Errorf("failed to stop otel: %w", err))
		}

		if err := redisClient.Close(); err != nil {
			errs = errors.Join(errs, fmt.Errorf("failed to stop redis: %w", err))
		}

		postgresClient.Close()

		return errs
	})

	if err := group.Wait(); err != nil {
		log.Fatalf("job exit reason: %v", err)
	}

	stop()
}

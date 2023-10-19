package rest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/cep21/circuit/v3"

	"github.com/chatbot-go/app/domain/erring"
	"github.com/chatbot-go/app/gateway/api/rest/response"
	"github.com/chatbot-go/app/library/resource"
	"github.com/chatbot-go/app/telemetry"
)

func HandleWithCircuit(circ *circuit.Circuit, route string, handler func(*http.Request) *response.Response) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		span := trace.SpanFromContext(req.Context())
		span.SetAttributes(telemetry.AttributesFromContext(req.Context())...)

		code, desc := codes.Ok, ""

		var (
			resp  *response.Response
			start time.Time
		)

		err := circ.Execute(req.Context(), func(ctx context.Context) error {
			start = time.Now()

			resp = handler(req.WithContext(ctx))

			// If the error is expected, we don't want to open the circuit breaker,
			// so we return a circuit.SimpleBadRequest error.
			if errors.Is(resp.InternalErr, erring.ErrExpected) {
				return circuit.SimpleBadRequest{Err: resp.InternalErr}
			}

			return resp.InternalErr
		}, nil)
		if err != nil {
			code, desc = codes.Error, err.Error()

			// If we have an error but not a response error, we can assume that
			// it is a circuit breaker error and we can replace the original
			// response.
			// When a circuit breaker timeout is triggered, the context reaches
			// its deadline and our application returns a context.DeadlineExceeded
			// error, so we need to handle it as a CB error.
			if resp == nil || resp.InternalErr == nil || errors.Is(err, context.DeadlineExceeded) {
				resp = handleCircuitBreakerErrorResponse(err)
			}

			// If the error is a circuit.SimpleBadRequest, it means it is an
			// expected error, so we want to handle the original error.
			var expected circuit.SimpleBadRequest

			asExpected := errors.As(err, &expected)
			if asExpected {
				code, desc, err = codes.Ok, "", expected.Cause()
			}

			logLevel := slog.LevelError
			if asExpected || resp.Status == http.StatusBadRequest {
				logLevel = slog.LevelWarn
			}

			if !resp.OmitLogs {
				logAttrs := logAttrs(req, resp, route, time.Since(start))
				slog.Log(req.Context(), logLevel, err.Error(), logAttrs...)
			}

			span.RecordError(err)
		}

		err = sendJSON(rw, resp.Status, resp.Payload, resp.Headers)
		if err != nil {
			code, desc = codes.Error, err.Error()
			span.RecordError(err)

			logAttrs := logAttrs(req, resp, route, time.Since(start))
			slog.ErrorContext(req.Context(), err.Error(), logAttrs...)
		}

		span.SetStatus(code, desc)
	}
}

func sendJSON(rw http.ResponseWriter, statusCode int, payload any, header map[string]string) error {
	for key, value := range header {
		rw.Header().Set(key, value)
	}

	if payload == nil {
		rw.WriteHeader(statusCode)

		return nil
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(statusCode)

	err := json.NewEncoder(rw).Encode(payload)
	if err != nil {
		return fmt.Errorf("send json encode: %w", err)
	}

	return nil
}

func handleCircuitBreakerErrorResponse(err error) *response.Response {
	if errors.Is(err, context.DeadlineExceeded) {
		return &response.Response{
			Status: http.StatusRequestTimeout,
			Payload: response.Error{
				Type:    string(resource.SrnErrorRequestTimeout),
				Code:    "circuit-breaker:request-timeout",
				Message: "timeout",
			},
			InternalErr: err,
		}
	}

	var cbErr circuit.Error

	if errors.As(err, &cbErr) {
		switch {
		case cbErr.CircuitOpen():
			return &response.Response{
				Status: http.StatusServiceUnavailable,
				Payload: response.Error{
					Type:    string(resource.SrnErrorServiceUnavailable),
					Code:    "circuit-breaker:service-unavailable",
					Message: cbErr.Error(),
				},
				InternalErr: err,
			}
		case cbErr.ConcurrencyLimitReached():
			return &response.Response{
				Status: http.StatusTooManyRequests,
				Payload: response.Error{
					Type:    string(resource.SrnErrorTooManyRequests),
					Code:    "circuit-breaker:too-many-requests",
					Message: cbErr.Error(),
				},
				InternalErr: err,
			}
		}
	}

	return response.InternalServerError(err)
}

func logAttrs(req *http.Request, resp *response.Response, route string, duration time.Duration) []any {
	attrs := make([]any, 0)

	for key, value := range resp.LogAttrs {
		attrs = append(attrs, slog.Any(key, value))
	}

	httpAttrs := []any{
		slog.Int("status_code", resp.Status),
		slog.String("status_text", http.StatusText(resp.Status)),
		slog.String("method", req.Method),
		slog.String("url", req.URL.String()),
		slog.String("route", route),
		slog.String("proto", req.Proto),
		slog.String("remote_addr", req.RemoteAddr),
	}

	if host := req.Header.Get("Host"); host != "" {
		httpAttrs = append(httpAttrs, slog.String("host", host))
	}

	if trueClientIP := req.Header.Get("True-Client-IP"); trueClientIP != "" {
		httpAttrs = append(httpAttrs, slog.String("true_client_ip", trueClientIP))
	}

	if forwardedFor := req.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		httpAttrs = append(httpAttrs, slog.String("forwarded_for", forwardedFor))
	}

	if realIP := req.Header.Get("X-Real-IP"); realIP != "" {
		httpAttrs = append(httpAttrs, slog.String("real_ip", realIP))
	}

	if userAgent := req.UserAgent(); userAgent != "" {
		httpAttrs = append(httpAttrs, slog.String("user_agent", userAgent))
	}

	httpAttrs = append(httpAttrs, slog.Duration("duration", duration))

	httpGroup := slog.Group("http", httpAttrs...)

	return append(attrs, httpGroup)
}

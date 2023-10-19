package middleware

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/chatbot-go/app/library/ctxkey"
)

// HeadersToContext apply HTTP headers value to the context.
// Copies request id, idempotency key and authorization key to context.
// Also propagates istio B3 headers from request to response.
func HeadersToContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		const (
			authorizationHeaderName = "authorization"
			requestIDHeaderName     = "x-request-id"
		)

		ctx := req.Context()

		// Copy the authorization header value to the context.
		if authHeader := req.Header.Get(authorizationHeaderName); authHeader != "" {
			ctx = ctxkey.PutAuthorizationHeader(ctx, authHeader)
		}

		// Copy the request id header value to the context.
		requestID := req.Header.Get(requestIDHeaderName)
		if requestID == "" {
			requestID = uuid.NewString()
		}
		ctx = ctxkey.PutRequestID(ctx, requestID)
		rw.Header().Set(requestIDHeaderName, requestID)

		next.ServeHTTP(rw, req.WithContext(ctx))
	})
}

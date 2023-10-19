package middleware

import (
	"net/http"

	"github.com/chatbot-go/app/gateway/client/twilio"
)

func TwilioAuth(twilioClient *twilio.Client) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			ctx := req.Context()

			// Store Twilio's request URL (the url of your webhook) as a variable
			url := "https://" + req.Host + req.RequestURI

			// Store the application/x-www-form-urlencoded params from Twilio's request as a variable
			// In practice, this MUST include all received parameters, not a
			// hardcoded list of parameters that you receive today. New parameters
			// may be added without notice.
			req.ParseForm()

			params := map[string]string{}

			for key, values := range req.PostForm {
				params[key] = values[0]
			}

			// Store the X-Twilio-Signature header attached to the request as a variable
			signature := req.Header.Get("X-Twilio-Signature")

			// Check if the incoming signature is valid for your application URL and the incoming parameters
			if ok := twilioClient.RequestValidator.Validate(url, params, signature); !ok {
				rw.WriteHeader(http.StatusUnauthorized)

				return
			}

			next.ServeHTTP(rw, req.WithContext(ctx))
		})
	}
}

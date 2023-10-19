package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/chatbot-go/app/domain/usecase"
	"github.com/chatbot-go/app/gateway/api/middleware"
	"github.com/chatbot-go/app/gateway/api/rest"
	"github.com/chatbot-go/app/gateway/api/rest/response"
	"github.com/chatbot-go/app/gateway/client/twilio"
)

const (
	WebhooksTwilioCommand = "webhooks-twilio"
	WebhooksTwilioPattern = "/webhooks/twilio"
)

func (h *Handler) WebhooksTwilioSetup(router chi.Router, twilioClient *twilio.Client) {
	circuit := h.circuitManager.MustCreateCircuit(WebhooksTwilioCommand)
	handler := rest.HandleWithCircuit(circuit, WebhooksTwilioPattern, h.WebhooksTwilio)

	router = router.With(middleware.TwilioAuth(twilioClient))

	router.Post(WebhooksTwilioPattern, handler)
}

func (h *Handler) WebhooksTwilio(req *http.Request) *response.Response {
	input := usecase.EnqueueTwilioWebhookInput{
		MessageBody: req.PostForm.Get("Body"),
		MessageSid:  req.PostForm.Get("MessageSid"),
		PhoneNumber: req.PostForm.Get("From"),
	}

	err := h.useCase.EnqueueTwilioWebhook(req.Context(), input)
	if err != nil {
		return response.InternalServerError(err)
	}

	return response.Accepted(nil)
}

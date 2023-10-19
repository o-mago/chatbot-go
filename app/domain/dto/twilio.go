package dto

import "github.com/chatbot-go/app/domain/types"

type WebhookTwilio struct {
	MessageSid  string `json:"message_sid"`
	MessageBody string `json:"message_body"`
	PhoneNumber string `json:"phone_number"`
}

type SendMessageTemplateInput struct {
	Provider
	DestinationNumber string
	TemplateID        types.TwilioTemplate
	Variables         map[string]string
}

type SendMessageInput struct {
	Provider
	DestinationNumber string
	Message           string
}

type Provider string

const WhatsappProvider Provider = "whatsapp"

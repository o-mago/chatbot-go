package twilio

import (
	"github.com/twilio/twilio-go"
	validatorClient "github.com/twilio/twilio-go/client"

	"github.com/chatbot-go/app/config"
)

type Client struct {
	client              *twilio.RestClient
	RequestValidator    validatorClient.RequestValidator
	originNumber        string
	messagingServiceSid string
}

func NewClient(twilioConfig config.Twilio) *Client {
	client := twilio.NewRestClient()

	requestValidator := validatorClient.NewRequestValidator(twilioConfig.AuthToken)

	return &Client{
		client:              client,
		RequestValidator:    requestValidator,
		originNumber:        twilioConfig.OriginNumber,
		messagingServiceSid: twilioConfig.MessagingServiceSid,
	}
}

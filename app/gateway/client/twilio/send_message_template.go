package twilio

import (
	"context"
	"encoding/json"
	"fmt"

	api "github.com/twilio/twilio-go/rest/api/v2010"

	"github.com/chatbot-go/app/domain/dto"
	"github.com/chatbot-go/app/domain/erring"
)

//nolint:revive
func (c *Client) SendMessageTemplate(ctx context.Context, input dto.SendMessageTemplateInput) error {
	const operation = "Client.Twilio.SendMessage"

	params := &api.CreateMessageParams{}
	params.SetFrom(string(input.Provider) + ":" + c.originNumber)
	params.SetTo(string(input.Provider) + ":" + input.DestinationNumber)
	params.SetMessagingServiceSid(c.messagingServiceSid)
	params.SetContentSid(string(input.TemplateID))

	if input.Variables != nil {
		mapVariables, err := json.Marshal(input.Variables)
		if err != nil {
			return fmt.Errorf("%s -> %w", operation, err)
		}

		params.SetContentVariables(string(mapVariables))
	}

	resp, err := c.client.Api.CreateMessage(params)
	if err != nil {
		return fmt.Errorf("%s -> %w", operation, err)
	}

	// TODO: try to validate a case where SID is empty
	if resp.Sid == nil {
		return fmt.Errorf("%s -> %w", operation, erring.ErrMissingTwilioSid)
	}

	return nil
}

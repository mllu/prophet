package slack

import (
	"fmt"
	"log"
	"strings"

	"github.com/mllu/prophet/pkg/diagflow"

	"github.com/mllu/slack"
)

const (
	// action is used for slack attament action.
	ActionSelect = "select"
	ActionStart  = "start"
	ActionCancel = "cancel"
)

type SlackListener struct {
	client    *slack.Client
	botID     string
	channelID string
}

func NewSlackListener(client *slack.Client, botID, channelID string) *SlackListener {
	return &SlackListener{
		client:    client,
		botID:     botID,
		channelID: channelID,
	}
}

// LstenAndResponse listens slack events and response
// particular messages. It replies by slack message button.
func (s *SlackListener) ListenAndResponse() {
	rtm := s.client.NewRTM()

	// Start listening slack events
	go rtm.ManageConnection()

	// Handle slack events
	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			if err := s.handleMessageEvent(ev, rtm); err != nil {
				log.Printf("[ERROR] Failed to handle message: %s", err)
			}
		}
	}
}

// handleMesageEvent handles message events.
func (s *SlackListener) handleMessageEvent(ev *slack.MessageEvent, rtm *slack.RTM) error {
	log.Println("handleMessageEvent begin", ev.Channel, ev.Text)
	// Only response in specific channel. Ignore else.
	/*
		if ev.Channel != s.channelID {
			log.Printf("%s %s", ev.Channel, ev.Msg.Text)
			return nil
		}
	*/

	// Only response mention to bot. Ignore else.
	if !strings.HasPrefix(ev.Text, fmt.Sprintf("<@%s> ", s.botID)) {
		return nil
	}
	log.Println("handleMessageEvent begin")

	// Parse message
	m := strings.Split(strings.TrimSpace(ev.Text), " ")[1:]
	if len(m) == 0 {
		return fmt.Errorf("invalid message %s", m)
	}

	query := strings.Join(m, " ")
	log.Printf("Query : %s", query)

	dialogFlowResponse := diagflow.GetResponse(query)
	log.Printf("Response : %v", dialogFlowResponse)
	if dialogFlowResponse.Action == "input.unknown" {
		s.sendAttachments(ev)
		return nil
	}

	rtm.SendMessage(rtm.NewOutgoingMessage(dialogFlowResponse.Fulfillment.Speech, ev.Channel))

	return nil
}

func (s *SlackListener) sendAttachments(ev *slack.MessageEvent) error {
	// value is passed to message handler when request is approved.
	attachment := slack.Attachment{
		Text:       "Which beer do you want? :beer:",
		Color:      "#f9a41b",
		CallbackID: "beer",
		Actions: []slack.AttachmentAction{
			{
				Name: ActionSelect,
				Type: "select",
				Options: []slack.AttachmentActionOption{
					{
						Text:  "Asahi Super Dry",
						Value: "Asahi Super Dry",
					},
					{
						Text:  "Kirin Lager Beer",
						Value: "Kirin Lager Beer",
					},
					{
						Text:  "Sapporo Black Label",
						Value: "Sapporo Black Label",
					},
					{
						Text:  "Suntory Malts",
						Value: "Suntory Malts",
					},
					{
						Text:  "Yona Yona Ale",
						Value: "Yona Yona Ale",
					},
				},
			},
			{
				Name:  ActionCancel,
				Text:  "Cancel",
				Type:  "button",
				Style: "danger",
			},
		},
	}

	params := slack.MsgOptionAttachments(attachment)

	if _, _, err := s.client.PostMessage(ev.Channel, slack.MsgOptionText("", false), params); err != nil {
		return fmt.Errorf("failed to post message: %s", err)
	}

	return nil

}

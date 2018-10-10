package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	mySlack "github.com/mllu/prophet/pkg/slack"

	"github.com/mllu/slack"
)

// interactionHandler handles interactive message response.
type interactionHandler struct {
	slackClient       *slack.Client
	verificationToken string
}

func NewInteractionHandler(client *slack.Client, verificationToken string) *interactionHandler {
	return &interactionHandler{
		slackClient:       client,
		verificationToken: verificationToken,
	}
}

func (h *interactionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("interactionHandler begin")
	if r.Method != http.MethodPost {
		log.Printf("[ERROR] Invalid method: %s", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("[ERROR] Failed to read request body: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsonStr, err := url.QueryUnescape(string(buf)[8:])
	if err != nil {
		log.Printf("[ERROR] Failed to unespace request body: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var message slack.InteractionCallback
	if err := json.Unmarshal([]byte(jsonStr), &message); err != nil {
		log.Printf("[ERROR] Failed to decode json message from slack: %s", jsonStr)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Printf("jsonStr: %s", jsonStr)

	// Only accept message from slack with valid token
	if message.Token != h.verificationToken {
		log.Printf("[ERROR] Invalid token: %s", message.Token)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	ts := message.OriginalMessage.Timestamp
	channel := message.Channel.ID

	action := message.Actions[0]
	switch action.Name {
	case mySlack.ActionSelect:
		value := action.SelectedOptions[0].Value

		// Update original drop down message.
		attachment := slack.Attachment{
			Text:       fmt.Sprintf("OK to order %s ?", strings.Title(value)),
			Color:      "#f9a41b",
			CallbackID: "beer",
			Actions: []slack.AttachmentAction{
				{
					Name:  mySlack.ActionStart,
					Text:  "Yes",
					Type:  "button",
					Value: "start",
					Style: "primary",
				},
				{
					Name:  mySlack.ActionCancel,
					Text:  "No",
					Type:  "button",
					Style: "danger",
				},
			},
		}
		params := slack.MsgOptionAttachments(attachment)
		log.Println("channel:", channel, "ts:", ts, "params:", params)
		_, _, _, err := h.slackClient.UpdateMessage(channel, ts, params)
		if err != nil {
			log.Printf("failed to post message: %s", err)
			return
		}
		return
	case mySlack.ActionStart:
		title := ":ok: your order was submitted! yay!"
		h.responseMessage(w, channel, ts, title, "")
		return
	case mySlack.ActionCancel:
		title := fmt.Sprintf(":x: @%s canceled the request", message.User.Name)
		h.responseMessage(w, channel, ts, title, "")
		return
	default:
		log.Printf("[ERROR] ]Invalid action was submitted: %s", action.Name)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// responseMessage response to the original slackbutton enabled message.
// It removes button and replace it with message which indicate how bot will work
func (h *interactionHandler) responseMessage(w http.ResponseWriter, channel, ts, title, value string) {
	attachment := slack.Attachment{
		Color:      "#f9a41b",
		CallbackID: "beer",
		Actions:    []slack.AttachmentAction{},
		Fields: []slack.AttachmentField{
			{
				Title: title,
				Value: value,
				Short: false,
			},
		},
	}
	params := slack.MsgOptionAttachments(attachment)
	log.Println("channel:", channel, "ts:", ts, "params:", params)
	_, _, _, err := h.slackClient.UpdateMessage(channel, ts, params)
	if err != nil {
		log.Printf("failed to post message: %s", err)
	}
}

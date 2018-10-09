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

	"github.com/nlopes/slack"
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

	var message slack.AttachmentActionCallback
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

	action := message.Actions[0]
	switch action.Name {
	case mySlack.ActionSelect:
		value := action.SelectedOptions[0].Value

		// Overwrite original drop down message.
		originalMessage := message.OriginalMessage
		originalMessage.Attachments[0].Text = fmt.Sprintf("OK to order %s ?", strings.Title(value))
		originalMessage.Attachments[0].Actions = []slack.AttachmentAction{
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
		}

		w.Header().Add("Content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(&originalMessage)
		return
	case mySlack.ActionStart:
		title := ":ok: your order was submitted! yay!"
		h.responseMessage(w, message, title, "")
		return
	case mySlack.ActionCancel:
		title := fmt.Sprintf(":x: @%s canceled the request", message.User.Name)
		h.responseMessage(w, message, title, "")
		return
	default:
		log.Printf("[ERROR] ]Invalid action was submitted: %s", action.Name)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// responseMessage response to the original slackbutton enabled message.
// It removes button and replace it with message which indicate how bot will work
func (h *interactionHandler) responseMessage(w http.ResponseWriter, message slack.AttachmentActionCallback, title, value string) {
	log.Printf("message: %v", message)
	log.Printf("OriginalMessage.Msg: %v", message.OriginalMessage.Msg)
	log.Printf("OriginalMessage.Msg.Attachments: %v", message.OriginalMessage.Msg.Attachments)
}

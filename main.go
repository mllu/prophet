package main

import (
	"log"
	"net/http"

	"github.com/mllu/prophet/pkg/config"
	"github.com/mllu/prophet/pkg/handler"
	mySlack "github.com/mllu/prophet/pkg/slack"

	"github.com/kelseyhightower/envconfig"
	"github.com/nlopes/slack"
)

func main() {

	var env config.EnvConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Printf("[ERROR] Failed to process env var: %s", err)
	}

	// Listening slack event and response
	log.Printf("[INFO] Start slack event listening")
	slackClient := slack.New(env.BotToken)
	slackListener := mySlack.NewSlackListener(slackClient, env.BotID, env.ChannelID)
	go slackListener.ListenAndResponse()

	// Register handler to receive interactive message
	// responses from slack (kicked by user action)
	http.Handle("/interaction", handler.NewInteractionHandler(slackClient, env.VerificationToken))

	http.HandleFunc("/", handler.ChallengeHandler)

	log.Printf("[INFO] Server listening on :%s", env.Port)
	if err := http.ListenAndServe(":"+env.Port, nil); err != nil {
		log.Printf("[ERROR] %s", err)
	}

}

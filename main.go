package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/mllu/prophet/pkg/config"
	"github.com/mllu/prophet/pkg/handler"
	mySlack "github.com/mllu/prophet/pkg/slack"

	"github.com/fvbock/endless"
	"github.com/gin-gonic/gin"
	"github.com/golang/protobuf/jsonpb"
	"github.com/kelseyhightower/envconfig"
	"github.com/mllu/slack"
	"github.com/sirupsen/logrus"
	"google.golang.org/genproto/googleapis/cloud/dialogflow/v2"
)

func handleWebhook(c *gin.Context) {
	var err error

	wr := dialogflow.WebhookRequest{}
	if err = jsonpb.Unmarshal(c.Request.Body, &wr); err != nil {
		logrus.WithError(err).Error("Couldn't Unmarshal request to jsonpb")
		c.Status(http.StatusBadRequest)
		return
	}
	log.Printf("%v", wr)
	fmt.Println("I got a webhook", wr.GetQueryResult().GetOutputContexts())
}

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

	/*
		// Register handler to receive interactive message
		// responses from slack (kicked by user action)
		http.Handle("/interaction", handler.NewInteractionHandler(slackClient, env.VerificationToken))
	*/

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
	mux.HandleFunc("/interaction", handler.NewInteractionHandler(slackClient, env.VerificationToken).ServeHTTP)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", env.Port),
		Handler: mux,
	}

	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt)

	go func() {
		log.Printf("[INFO] Default Server listening on :%s", env.Port)
		if err := srv.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				log.Fatal(err)
			}
		}
	}()

	r := gin.Default()
	r.POST("/webhook", handleWebhook)

	log.Printf("[INFO] Gin Server listening on :%s", ":8080")
	endless.ListenAndServe(":8080", r)
	/*
		if err = r.Run(":8080"); err != nil {
			logrus.WithError(err).Fatal("Couldn't start server")
		}
	*/

	<-stop

	log.Println("shutting down ...\n")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}

	log.Println("Bye...\n")
}

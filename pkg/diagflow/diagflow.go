package diagflow

import (
	"log"
	"os"

	dfgc "github.com/mllu/dialogflow-go-client"
	apiai "github.com/mllu/dialogflow-go-client/models"
)

// GetResponse query diagflow to get responses
func GetResponse(input string) apiai.Result {
	err, client := dfgc.NewDialogFlowClient(apiai.Options{
		AccessToken: os.Getenv("DIALOG_FLOW_TOKEN"),
	})
	if err != nil {
		log.Fatal(err)
	}

	query := apiai.Query{
		Query: input,
	}
	resp, err := client.QueryFindRequest(query)
	if err != nil {
		log.Fatal(err)
	}
	return resp.Result
}

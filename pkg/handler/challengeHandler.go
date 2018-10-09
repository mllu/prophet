package handler

import (
	"encoding/json"
	"log"
	"net/http"
)

type data struct {
	Category  string `json:"type"`
	Token     string `json:"token"`
	Challenge string `json:"challenge"`
}

type challenge struct {
	Challenge string `json:"challenge"`
}

func ChallengeHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("challengeHandler")
	cObj := &data{}

	err := json.NewDecoder(r.Body).Decode(cObj)
	if err != nil {
		log.Printf("[ERROR] %s", err)
		panic(err)
	}

	log.Printf("[CHALLENGE] %s", cObj.Challenge)
	resp := &challenge{}
	resp.Challenge = cObj.Challenge
	cJSON, err := json.Marshal(resp)
	if err != nil {
		panic(err)
	}

	//set content-type to json
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(cJSON)
}

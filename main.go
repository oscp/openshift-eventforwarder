package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"golang.org/x/build/kubernetes/api"
	"log"
	"net/http"
	"os"
	"time"
)

type Stream struct {
	Type  string    `json:"type,omitempty"`
	Event api.Event `json:"object"`
}

func main() {
	apiAddr := os.Getenv("OPENSHIFT_API_URL")
	apiToken := os.Getenv("OPENSHIFT_TOKEN")

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("GET", apiAddr+"/api/v1/events?watch=true", nil)
	if err != nil {
		log.Fatal("## Error while opening connection to openshift api", err)
	}
	req.Header.Add("Authorization", "Bearer "+apiToken)

	for {
		resp, err := client.Do(req)

		if err != nil {
			log.Println("## Error while connecting to:", apiAddr, err)
			time.Sleep(5 * time.Second)
			break
		}

		reader := bufio.NewReader(resp.Body)

		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				log.Println("## Error reading from response stream.", err)
				break
			}

			event := Stream{}
			decErr := json.Unmarshal(line, &event)
			if decErr != nil {
				log.Println("## Error decoding json", err)
				break
			}

			fmt.Printf("%v | Project: %v | Name: %v | Kind: %v | Reason: %v | Message: %v\n",
				event.Event.LastTimestamp,
				event.Event.Namespace, event.Event.Name,
				event.Event.Kind, event.Event.Reason, event.Event.Message)
		}
	}
}


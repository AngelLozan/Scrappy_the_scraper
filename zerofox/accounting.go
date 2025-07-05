package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Alert struct {
	ID        int    `json:"id"`
	Url       string `json:"offending_content_url"`
	Timestamp string `json:"content_created_at"`
	Status    string `json:"status"`
	Severity  int    `json:"severity"`
}

type Alerts struct {
	Alerts []Alert `json:"alerts"`
}

func fetchAlerts() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	t := time.Now()
	maxmax_timestamp := t.Format(time.RFC3339)
	zfToken := os.Getenv("ZF_TOKEN")
	params := url.Values{}
	params.Add("max_timestamp", maxmax_timestamp)
	params.Add("status", "takedown_requested")
	params.Add("status", "takedown_submitted")
	req, err := http.NewRequest("GET", "https://api.zerofox.com/1.0/alerts?"+params.Encode(), nil)
	fmt.Println("Request URL:", req.URL.String())
	if err != nil {
		log.Printf("Failed to create request: %s", err)
		return
	}
	req.Header.Add("Authorization", zfToken)
	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		log.Printf("Request Failed: %s", err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	// Log the request body
	bodyString := string(body)
	log.Print(bodyString)
	// Unmarshal result
	alerts := Alerts{}
	err = json.Unmarshal(body, &alerts)
	if err != nil {
		log.Printf("Reading body failed: %s", err)
		return
	}

	log.Printf("Length of alerts: %d", len(alerts.Alerts))

}

func main() {
	fetchAlerts()
}

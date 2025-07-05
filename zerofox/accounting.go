package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
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

// Fetch and format alerts urls
func fetchAlerts() []Alert {
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
		return []Alert{}
	}
	req.Header.Add("Authorization", zfToken)
	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		log.Printf("Request Failed: %s", err)
		return []Alert{}
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
		return []Alert{}
	}

	log.Printf("Length of alerts: %d", len(alerts.Alerts))
	onlyAlerts := alerts.Alerts[:0] // Create a slice to hold only the alerts
	for _, alert := range alerts.Alerts {
		alert.Url = cleanUrl(alert.Url)
		if alert.Url == "" {
			continue
		} else {
			onlyAlerts = append(onlyAlerts, alert)
		}
	}

	if len(onlyAlerts) > 0 {
		fmt.Println("Found", len(onlyAlerts), "alerts")
		return onlyAlerts
	} else {
		fmt.Println("No alerts found")
		return []Alert{}
	}
}

func cleanUrl(rawUrl string) string {
	cleaned := strings.TrimPrefix(rawUrl, "hxxp://")

	parsed, err := url.Parse("http://" + cleaned)
	if err != nil {
		fmt.Println("Error parsing URL:", err)
		return ""
	}

	host := parsed.Host
	fmt.Println("Final domain:", host)
	return host
}

// TO DO:
// For each alert url, check on hitlist if it is resolved status
// if resolved, cancel takedown and close alert in zerofox

func main() {
	fetchAlerts()
}

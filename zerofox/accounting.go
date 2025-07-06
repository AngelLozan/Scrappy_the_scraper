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

	"github.com/AngelLozan/scraper/types"
	"github.com/joho/godotenv"
)

// Fetch and format alerts urls
func fetchAlerts() []types.Alert {
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
	fmt.Println("\n Request URL: ", req.URL.String())
	if err != nil {
		log.Printf("Failed to create request: %s", err)
		return []types.Alert{}
	}
	req.Header.Add("Authorization", zfToken)
	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		log.Printf("Request Failed: %s", err)
		return []types.Alert{}
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	bodyString := string(body)
	log.Print(bodyString)
	// Unmarshal result
	alerts := types.Alerts{}
	err = json.Unmarshal(body, &alerts)
	if err != nil {
		log.Printf("Reading body failed: %s", err)
		return []types.Alert{}
	}

	log.Printf("\n\n Length of alerts: %d \n\n", len(alerts.Alerts))
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
		fmt.Println("❌ No alerts found")
		return []types.Alert{}
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

// For each alert url, check on hitlist if it is resolved status
func reconcileHitlist(alerts []types.Alert) {
	for _, alert := range alerts {
		fmt.Println("Checking hitlist for alert:", alert.Url)
		if searchHitlist(alert.Url) {
			fmt.Println("\n\n ✅ Alert found in hitlist, cancelling takedown and closing alert")
			cancelTakedown(alert)
		} else {
			fmt.Println("\n\n ❌ Alert not found in hitlist, no action taken")
			continue
		}
	}
}

// If resolved, cancel takedown and close alert in zerofox
func cancelTakedown(alert types.Alert) {
	fmt.Println("\n\n 🛠️Cancelling takedown for alert:", alert.Url)

	zfToken := os.Getenv("ZF_TOKEN")
	requestUrl := fmt.Sprintf("https://api.zerofox.com/1.0/alerts/%d/cancel_takedown", alert.ID)
	req, err := http.NewRequest("POST", requestUrl, nil)
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

	bodyString := string(body)
	log.Print(bodyString)

	if err != nil {
		log.Printf("Reading body failed: %s", err)
		return
	}

	fmt.Println("/n/n ✅ Takedown cancelled for alert:", alert.Url)
	closed, err := closeAlert(alert)
	if err != nil {
		log.Printf("Failed to close alert: %s", err)
		return
	}
	if closed {
		fmt.Println("\n\n ✅ Alert closed successfully:", alert.Url)
	}

}

func closeAlert(alert types.Alert) (bool, error) {
	fmt.Println("Closing alert:", alert.Url)

	zfToken := os.Getenv("ZF_TOKEN")
	requestUrl := fmt.Sprintf("https://api.zerofox.com/1.0/alerts/%d/close", alert.ID)
	req, err := http.NewRequest("POST", requestUrl, nil)
	fmt.Println(" \n\n Request URL:", req.URL.String())
	if err != nil {
		log.Printf("Failed to create request: %s", err)
		return false, err
	}
	req.Header.Add("Authorization", zfToken)
	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		log.Printf("Request Failed: %s", err)
		return false, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	bodyString := string(body)
	log.Print(bodyString)

	if err != nil {
		log.Printf("Reading body failed: %s", err)
		return false, err
	}

	if resp.StatusCode == http.StatusOK {
		fmt.Println("\n\n ✅ Alert closed successfully:", alert.Url)
		return true, nil
	}
	
	return false, fmt.Errorf("failed to close alert: %s", alert.Url)
}

func searchHitlist(_url string) bool {
	fmt.Println("\n\n 🔎 Searching hitlist for URL:", _url)
	params := url.Values{}
	params.Add("q", _url)
	resp, err := http.Get("https://scam-hitlist-p.ot.exodus.com/api/iocs/search/?" + params.Encode())
	if err != nil {
		log.Printf("Request Failed: %s", err)
		return false
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	bodyString := string(body)
	log.Print(bodyString)

	zf_alerts := types.HitlistAlerts{}
	err = json.Unmarshal(body, &zf_alerts)
	if err != nil {
		log.Printf("Reading body failed: %s", err)
		return false
	}

	if len(zf_alerts.HitlistAlerts) > 0 && zf_alerts.HitlistAlerts[0].Status == "resolved" {
		fmt.Println("\n\n 🟩 Found", len(zf_alerts.HitlistAlerts), "alert is resolved in hitlist")
		return true
	} else {
		fmt.Println("/n/n ❌ No alerts found in hitlist")
		return false
	}
}

func main() {
	alerts := fetchAlerts()
	reconcileHitlist(alerts)
}

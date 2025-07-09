package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
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
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Reading body failed: %s", err)
		return []types.Alert{}
	}

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
func reconcileHitlist(alerts []types.Alert) int {
	count := 0
	for _, alert := range alerts {
		fmt.Println("Checking hitlist for alert:", alert.Url)
		if searchHitlist("./zerofox/data.csv", alert.Url) {
			fmt.Println("\n\n ✅ Alert found in hitlist, cancelling takedown and closing alert")
			cancelTakedown(alert)
			count++
		} else {
			fmt.Println("\n\n ❌ Alert not found in hitlist, no action taken")
			continue
		}
	}
	fmt.Println("\n\n 🏁 Reconciliation complete, total alerts processed:", count)
	return count
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
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Reading body failed: %s", err)
		return
	}

	bodyString := string(body)
	log.Print(bodyString)

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
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Reading body failed: %s", err)
		return false, err
	}

	bodyString := string(body)
	log.Print(bodyString)

	if resp.StatusCode == http.StatusOK {
		fmt.Println("\n\n ✅ Alert closed successfully:", alert.Url)
		return true, nil
	}

	return false, fmt.Errorf("failed to close alert: %s", alert.Url)
}

type IOC struct {
	URL      string
	Resolved bool
}

func loadIOCsFromCSV(path string) ([]IOC, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	csvReader := csv.NewReader(file)
	csvReader.TrimLeadingSpace = true
	csvReader.FieldsPerRecord = -1 // allow variable columns

	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) < 1 {
		return nil, fmt.Errorf("CSV is empty or missing headers")
	}

	header := records[0]
	iocIndex := -1
	statusIndex := -1

	for i, h := range header {
		switch strings.TrimSpace(h) {
		case "IOC":
			iocIndex = i
		case "Status":
			statusIndex = i
		}
	}

	if iocIndex == -1 || statusIndex == -1 {
		return nil, fmt.Errorf("required columns not found")
	}

	var iocs []IOC
	for _, line := range records[1:] {
		if len(line) <= statusIndex || len(line) <= iocIndex {
			continue
		}

		status := strings.ToLower(strings.TrimSpace(line[statusIndex]))
		url := strings.TrimSpace(line[iocIndex])
		resolved := status == "resolved"

		iocs = append(iocs, IOC{
			URL:      url,
			Resolved: resolved,
		})
	}

	return iocs, nil
}


func searchHitlist(path string, _url string) bool {
	iocs, err := loadIOCsFromCSV(path)
	if err != nil {
		log.Fatalf("Failed to load CSV: %v", err)
	}

	for _, ioc := range iocs {
		fmt.Printf("Checking IOC: %s, Resolved: %t\n", ioc.URL, ioc.Resolved)
		if ioc.Resolved && strings.Contains(ioc.URL, _url) {
			log.Printf("✅ Found resolved alert for URL: %s", _url)
			return true
		}
	}
	log.Printf("❌ No resolved alert for URL: %s", _url)
	return false
}

func main() {
	alerts := fetchAlerts()
	reconcileHitlist(alerts)
}

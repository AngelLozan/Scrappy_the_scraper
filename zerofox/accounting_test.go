package main

import (
	"testing"
	"encoding/json"
	"os"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/AngelLozan/scraper/types"
)

func TestFetchAlerts(t *testing.T) {
	alerts, err := os.ReadFile("accounting_test_data.json")
	if err != nil {
		t.Fatalf("Failed to read JSON file: %v", err)
	}

	var alertsData types.Alerts
	err = json.Unmarshal(alerts, &alertsData)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON data: %v", err)
	}
	

	for _, alert := range alertsData.Alerts {
		if alert.ID == 0 || alert.Url == "" || alert.Timestamp == "" || alert.Status == "" {
			t.Errorf("Alert has missing fields: %+v", alert)
		}
	}
}

func closeAlertTest(_url, base string) (types.Alert, bool) {
	resp, err := http.Get(base + "?q=" + _url)
	if err != nil {
		return types.Alert{}, false
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var alerts []types.Alert
	err = json.Unmarshal(body, &alerts)
	if err != nil || len(alerts) == 0 {
		return types.Alert{}, false
	}
	return alerts[0], alerts[0].Status == "closed"
}

func TestCloseAlert(t *testing.T) {
	mockData := `[{"id":16299,"offending_content_url":"https://lyumlabs-v2connect.pages.dev/Connect","content_created_at":"2023-10-01T12:00:00Z","status":"closed","severity":1}]`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockData))
	}))
	defer ts.Close()

	urlToClose := "https://lyumlabs-v2connect.pages.dev/Connect"
	alert, ok := closeAlertTest(urlToClose, ts.URL)
	if !ok {
		t.Fatalf("Expected status 'closed', got failure")
	}
	if alert.Url != urlToClose {
		t.Errorf("Expected URL %s, got %s", urlToClose, alert.Url)
	}
	if alert.Status != "closed" {
		t.Errorf("Expected status 'closed', got %s", alert.Status)
	}
}

func searchHitlistTest(_url, base string) (types.HitlistAlert, bool) {
	resp, err := http.Get(base + "?q=" + _url)
	if err != nil {
		return types.HitlistAlert{}, false
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var alerts []types.HitlistAlert
	err = json.Unmarshal(body, &alerts)
	if err != nil || len(alerts) == 0 {
		return types.HitlistAlert{}, false
	}
	return alerts[0], alerts[0].Status == "resolved"
}

func TestSearchHitlist_Unmarshal(t *testing.T) {
	// Mock JSON response
	mockData := `[{"id":16299,"url":"https://lyumlabs-v2connect.pages.dev/Connect","status":"resolved","tags":["wallet_connection "]}]`

	// Create a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockData))
	}))
	defer ts.Close()

	urlToCheck := "https://lyumlabs-v2connect.pages.dev/Connect"
	alert, ok := searchHitlistTest(urlToCheck, ts.URL)

	if !ok {
		t.Fatalf("Expected status 'resolved', got failure")
	}
	if alert.Url != urlToCheck {
		t.Errorf("Expected URL %s, got %s", urlToCheck, alert.Url)
	}
}
package types

import (
	"time"
)

type Malware struct {
	Link  string
	Title string
}

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

type HitlistAlert struct {
	ID          int       `json:"id"`
	Url         string    `json:"url"`
	RemovedDate *time.Time `json:"removed_date"`
	Status      string    `json:"status"`
}

type HitlistAlerts struct {
	HitlistAlerts []HitlistAlert 
}

type Snap struct {
	Title       string `json:"title"`
	Summary     string `json:"summary"`
	Description string `json:"description"`
	PackageName string `json:"package_name"`
	Publisher   string `json:"publisher"`
	Website     string `json:"website"`
}

type Embedded struct {
	Packages []Snap `json:"clickindex:package"`
}

type SnapAPIResponse struct {
	Embedded Embedded `json:"_embedded"`
}
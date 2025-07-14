package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strings"

	"github.com/AngelLozan/scraper/types"
	"github.com/joho/godotenv"
)

func sendEmail(items []types.Malware) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	emailAppPassword := os.Getenv("APP_PASS")
	yourMail := os.Getenv("SENDER")
	recipient := os.Getenv("RECIPIENT")
	hostAddress := "smtp.gmail.com"

	authenticate := smtp.PlainAuth("", yourMail, emailAppPassword, hostAddress)
	// tlsConfigurations := &tls.Config{
	// 	InsecureSkipVerify: true,
	// 	ServerName:         hostAddress,
	// }

	var body string
	for _, item := range items {
		body += fmt.Sprintf("%v: %v\n\n", item.Title, item.Link)
	}
	to := []string{recipient}

	msg := []byte(fmt.Sprintf("To: %v\r\n"+

		"Subject: Malicious packages found on Snap\r\n"+

		"\r\n"+

		"Please review the following packages: \n%v\r\n", recipient, body))

	error := smtp.SendMail("smtp.gmail.com:587", authenticate, yourMail, to, msg)

	if error != nil {

		log.Fatal(error)
	}

	fmt.Println("Successful, the mail was sent!")

}

func scrapeSnapAPI() {
	url := "https://api.snapcraft.io/api/v1/snaps/search?q=exodus"
	var items []types.Malware
	maliciousItem := types.Malware{}

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Snap-Device-Series", "16")
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal("Request failed:", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	log.Println("Response Status:", resp.Status)
	// log.Println("Response Body:", string(body))

	var apiResp types.SnapAPIResponse

	if err := json.Unmarshal(body, &apiResp); err != nil {
		log.Fatal("Failed to parse JSON:", err)
	}

	keywords := []string{"exodus", "crypto", "wallet"}
	for _, snap := range apiResp.Embedded.Packages {
		text := strings.ToLower(snap.Title + " " + snap.Summary)
		for _, kw := range keywords {
			if strings.Contains(text, kw) && !strings.Contains(text, "kubelogin") {
				fmt.Printf("Matched: %s - %s\n", snap.Title, snap.Summary)
				maliciousItem.Title = snap.Title
				maliciousItem.Link = "https://snapcraft.io/" + snap.PackageName
				items = append(items, maliciousItem)
				maliciousItem = types.Malware{}
				break
			}
		}
	}
	if len(items) > 0 {
		fmt.Println("Found", len(items), "malicious items")
		sendEmail(items)
	} else {
		fmt.Println("❌ No malicious items found")
	}
}

func main() {
	scrapeSnapAPI()
	fmt.Println("Scraping completed.")
}

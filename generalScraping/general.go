package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/smtp"
	"net/url"
	"os"
	"strings"

	"github.com/AngelLozan/scraper/types"
	"github.com/gocolly/colly/v2"
	"github.com/joho/godotenv"
)

func sendPotentialMaliciousEmail(items []types.Malware) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	emailAppPassword := os.Getenv("APP_PASS")
	yourMail := os.Getenv("SENDER")
	recipient := os.Getenv("RECIPIENT")
	hostAddress := "smtp.gmail.com"

	authenticate := smtp.PlainAuth("", yourMail, emailAppPassword, hostAddress)

	var body string
	for _, item := range items {
		body += fmt.Sprintf("%v: %v\n\n", item.Title, item.Link)
	}
	to := []string{recipient}

	msg := []byte(fmt.Sprintf("To: %v\r\n"+

		"Subject: Potentially Malicious Urls\r\n"+

		"\r\n"+

		"Please review the following sites: \n%v\r\n", recipient, body))

	error := smtp.SendMail("smtp.gmail.com:587", authenticate, yourMail, to, msg)

	if error != nil {

		log.Fatal(error)
	}

	fmt.Println("Successful, the mail was sent!")

}

func scrapeGeneral() {
	c := colly.NewCollector()
	var items []types.Malware

	element := "a.tilk"

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	c.OnResponse(func(r *colly.Response) {
		fmt.Printf("Got a response from %v\n\n", r.Request.URL)
	})

	c.OnError(func(r *colly.Response, e error) {
		fmt.Println("An error occurred!:", e)
	})

	c.OnHTML(element, func(e *colly.HTMLElement) {

		maliciousItem := types.Malware{}

		link := extractBingLink(e.Attr("href"))

		title := e.Attr("aria-label")

		cleanLink := strings.TrimSpace(link)
		cleanTitle := strings.TrimSpace(title)

		maliciousItem.Link = cleanLink
		maliciousItem.Title = cleanTitle

		excludedWords := []string{"www.exodus.com", "reddit"}
		shouldAppend := true
		for _, word := range excludedWords {
			if strings.Contains(strings.ToLower(cleanLink), word) {
				shouldAppend = false
				break
			}
		}

		if shouldAppend {
			items = append(items, maliciousItem)
		}

	})

	c.OnScraped(func(r *colly.Response) {
		fmt.Println("Scraping finished for", r.Request.URL)
	})

	pages := []int{1, 11, 21, 31, 41, 51, 61, 71, 81, 91}

	for _, start := range pages {
		searchURL := fmt.Sprintf("https://www.bing.com/search?q=exodus+wallet&first=%d", start)
		err := c.Visit(searchURL)
		if err != nil {
			log.Println("Visit failed for", searchURL, ":", err)

		}
	}

	c.Wait()

	if len(items) > 0 {
		fmt.Println("Found some potentially malicious items:", items)
		sendPotentialMaliciousEmail(items)
	} else {
		fmt.Println("Nothing found today")
	}

}

func extractBingLink(rawHref string) string {
	u, err := url.Parse(rawHref)
	if err != nil {
		log.Println("Invalid href:", err)
		return rawHref
	}

	encoded := u.Query().Get("u")
	if encoded == "" {
		return rawHref
	}

	cleanEncoded := encoded
	if strings.HasPrefix(encoded, "a1") {
		cleanEncoded = encoded[2:]
	}

	decodedBytes, err := base64.RawURLEncoding.DecodeString(cleanEncoded)
	if err != nil {
		log.Println("Base64 decode failed:", err)
		return rawHref
	}

	return string(decodedBytes)
}

func main() {
	scrapeGeneral()
	fmt.Println("Scraping completed.")
}

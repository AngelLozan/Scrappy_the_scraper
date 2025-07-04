package main

import (
	"fmt"
	"log"
	"net/smtp"
	// "os"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/joho/godotenv"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/AngelLozan/scraper/types"
)

func sendLambdaEmail(items []types.Malware) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// emailAppPassword := os.Getenv("APP_PASS")
	// yourMail := os.Getenv("SENDER")
	// recipient := os.Getenv("RECIPIENT")
	emailAppPassword := "gkki jurz prqd jnii"
	yourMail := "scott.lo@exodus.io"
	recipient := "scott.lo@exodus.io"
	hostAddress := "smtp.gmail.com"

	authenticate := smtp.PlainAuth("", yourMail, emailAppPassword, hostAddress)
	// tlsConfigurations := &tls.Config{
	// 	InsecureSkipVerify: true,
	// 	ServerName:         hostAddress,
	// }

	var body string
	for _, item := range items{
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

func scrapeLambda(){
	c := colly.NewCollector()

	url := "https://snapcraft.io/search?q=exodus"

	element := ".p-media-object"

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	c.OnResponse(func(r *colly.Response) {
		fmt.Printf("Got a response from %v\n\n", r.Request.URL)
	})

	c.OnError(func(r *colly.Response, e error) {
		fmt.Println("An error occurred!:", e)
	})

	var items []types.Malware

	c.OnHTML(element, func(e *colly.HTMLElement) {

		maliciousItem := types.Malware{}

		link := e.Attr("href")
		title := e.Attr("title")

		cleanLink := strings.TrimSpace(link)
		cleanTitle := strings.TrimSpace(title)

		maliciousItem.Link = fmt.Sprintf("https://snapcraft.io%s", cleanLink)
		maliciousItem.Title = cleanTitle

		if strings.Contains(strings.ToLower(maliciousItem.Title), "wallet") {
			items = append(items, maliciousItem)
		}

	})

		c.OnScraped(func(r *colly.Response) {
		if len(items) > 0 {
			fmt.Println("Found some malicious items:", items)
			sendLambdaEmail(items)
		} else {
			fmt.Println("Nothing found today")
		}
	})

	err := c.Visit(url)

	if err != nil {
		log.Fatal(err)
	}
}

type scrapeData struct {
	Name string `json:"name"`
}


func HandleRequest(event *scrapeData) (*string, error) {
	if event == nil {
		return nil, fmt.Errorf("received nil event")
	}
	message := fmt.Sprintf("Hello %s!", event.Name)
	go scrape()
	return &message, nil
}

func mainLambda() {
	lambda.Start(HandleRequest)
}
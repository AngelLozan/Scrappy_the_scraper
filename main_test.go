    package main

		import (
			"testing"
			"github.com/gocolly/colly/v2"
		)

		func TestVisitURLAndLogVisit(t *testing.T) {

			c := colly.NewCollector()

			url := "https://snapcraft.io/search?q=exodus"

			var visitedURL string

			c.OnRequest(func(r *colly.Request) {
				visitedURL = r.URL.String()
			})

			err := c.Visit(url)
			if err != nil {
				t.Fatal(err)
			}

			expectedURL := "https://snapcraft.io/search?q=exodus"
			if visitedURL != expectedURL {
				t.Errorf("Visited URL does not match expected URL. Got: %s, Expected: %s", visitedURL, expectedURL)
			}
		}


func TestInvalidURLAndLogError(t *testing.T) {

  c := colly.NewCollector()

  url := "https://invalidurl"

  var errorMessage string

  c.OnError(func(r *colly.Response, e error) {
    errorMessage = e.Error()
  })

  err := c.Visit(url)

  if err == nil {
    t.Error("Expected an error to occur, but got nil")
  }

  expectedErrorMessage := "Get \"https://invalidurl\": dial tcp: lookup invalidurl: no such host"
  if errorMessage != expectedErrorMessage {
    t.Errorf("Error message does not match expected error message. Got: %s, Expected: %s", errorMessage, expectedErrorMessage)
  }
}

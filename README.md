# Scrape snap store for malicious packages

A precursor function to be modified and run as a lambda in aws that will alert brand protection agents of malicious packages on the snap store, which are impersonating our brand. 

Beta (currently running cron job on my computer)

cronmain runs a scrape of the snap store
crongeneral runs a scrape of bing for general malicious sites

Ideas: auto-submit to hitlist. Consolidate packages and main.

### Setup

Build for local scraping:

- go build -o general general.go


Build the Go script and zip for AWS Lambda.
```
<!-- > GOOS=linux GOARCH=amd64 go build -o main main.go -->
GOARCH=arm64 GOOS=linux go build -o bootstrap lambda.go
<!-- > zip main.zip main -->
zip boostrap.zip boostrap
```

For boostrap test, set upload as zip and enter test event to follow struct of `scrapeData` Test should succeed. 

Needs `.env` vars initialized in AWS. 

Set hanlder to `main`
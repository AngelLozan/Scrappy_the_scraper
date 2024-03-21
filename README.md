# Scrape snap store for malicious packages

A function to be run as a lambda in aws that will alert brand protection agents of malicious packages on the snap store, which are impersonating our brand.

Beta

### Setup

Build the Go script and zip for AWS Lambda.
```
> GOOS=linux GOARCH=amd64 go build -o main main.go
> zip main.zip main
```

Needs `.env` vars initialized in AWS. 
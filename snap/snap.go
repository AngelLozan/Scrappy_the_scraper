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
	"unicode"

	"github.com/AngelLozan/scraper/types"
	"github.com/joho/godotenv"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
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

// homoglyphMap maps common homoglyph characters to their ASCII equivalents
var homoglyphMap = map[rune]rune{
	// Cyrillic homoglyphs (lowercase)
	'а': 'a', 'е': 'e', 'о': 'o', 'р': 'p', 'с': 'c', 'у': 'y', 'х': 'x',
	'ѕ': 's', 'і': 'i', 'ј': 'j', 'ԁ': 'd', 'ԍ': 'g', 'ԛ': 'q',
	// Cyrillic homoglyphs (uppercase)
	'А': 'A', 'В': 'B', 'Е': 'E', 'К': 'K', 'М': 'M', 'Н': 'H', 'О': 'O',
	'Р': 'P', 'С': 'C', 'Т': 'T', 'Х': 'X', 'Ѕ': 'S', 'І': 'I', 'Ј': 'J',
	// Greek homoglyphs (lowercase)
	'α': 'a', 'ο': 'o', 'ν': 'v', 'ι': 'i', 'υ': 'u', 'χ': 'x',
	'ε': 'e', 'η': 'n', 'κ': 'k', 'ρ': 'p', 'τ': 't', 'ω': 'w',
	// Greek homoglyphs (uppercase)
	'Α': 'A', 'Β': 'B', 'Ε': 'E', 'Ζ': 'Z', 'Η': 'H', 'Ι': 'I', 'Κ': 'K',
	'Μ': 'M', 'Ν': 'N', 'Ο': 'O', 'Ρ': 'P', 'Τ': 'T', 'Υ': 'Y', 'Χ': 'X',
	// Additional lookalikes
	'ⅰ': 'i', 'ⅼ': 'l', 'ⅾ': 'd', 'ⅿ': 'm',
	'０': '0', '１': '1', '２': '2', '３': '3', '４': '4',
	'５': '5', '６': '6', '７': '7', '８': '8', '９': '9',
	// Mathematical and other variants
	'ℯ': 'e', '℘': 'p', 'ℴ': 'o',
	// Zero-width and combining characters are handled separately
}

// normalizeAndRemoveHomoglyphs converts text to lowercase, normalizes unicode,
// and replaces common homoglyphs with their ASCII equivalents
func normalizeAndRemoveHomoglyphs(text string) string {
	// First, normalize unicode using NFD (Canonical Decomposition)
	// This separates base characters from combining marks
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	normalized, _, _ := transform.String(t, text)

	// Convert to lowercase
	normalized = strings.ToLower(normalized)

	// Replace homoglyphs
	var result strings.Builder
	for _, r := range normalized {
		if replacement, exists := homoglyphMap[r]; exists {
			result.WriteRune(replacement)
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// containsKeywordVariant checks if text contains a keyword or its homoglyph variants
func containsKeywordVariant(text, keyword string) bool {
	normalizedText := normalizeAndRemoveHomoglyphs(text)
	normalizedKeyword := strings.ToLower(keyword)
	return strings.Contains(normalizedText, normalizedKeyword)
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
		text := snap.Title + " " + snap.Summary
		// Check if text contains "kubelogin" (known false positive)
		if containsKeywordVariant(text, "kubelogin") {
			continue
		}
		for _, kw := range keywords {
			if containsKeywordVariant(text, kw) {
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

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

type Transaction struct {
	AccountID               int64     `json:"accountId"`
	Date                    time.Time `json:"date"`
	BankReferenceNumber     *string   `json:"bankReferenceNumber,omitempty"`
	Justification           *string   `json:"justification,omitempty"`
	Indicator               *string   `json:"indicator,omitempty"`
	MerchantIdentifier      *string   `json:"merchantIdentifier,omitempty"`
	Amount1                 *string   `json:"amount1,omitempty"`
	MCCCode                 *string   `json:"mccCode,omitempty"`
	CardMasked              *string   `json:"cardMasked,omitempty"`
	Reference               *string   `json:"reference,omitempty"`
	Description             *string   `json:"description,omitempty"`
	PaymentMethod           *string   `json:"paymentMethod,omitempty"`
	Direction               string    `json:"direction"`
	Amount                  string    `json:"amount"`
	BalanceAfterTransaction *string   `json:"balanceAfterTransaction,omitempty"`
	StatementFileName       *string   `json:"statementFileName,omitempty"`

	// Proposed additions
	RawText       string `json:"rawText,omitempty"`
	Currency      string `json:"currency,omitempty"`
	ParserVersion string `json:"parserVersion,omitempty"`
	BankCode      string `json:"bankCode,omitempty"`
}

func main() {
	if len(os.Args) < 2 {
		panic("usage: go run main.go statement.pdf")
	}

	pdfPath := os.Args[1]

	text, err := extractTextWithPdftotext(pdfPath)
	if err != nil {
		panic(err)
	}

	txs := parseTransactions(text, pdfPath)

	out, err := json.MarshalIndent(txs, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(out))
}

func extractTextWithPdftotext(pdfPath string) (string, error) {
	cmd := exec.Command("pdftotext", "-layout", pdfPath, "-")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("pdftotext failed: %w", err)
	}
	return string(out), nil
}

func parseTransactions(text string, fileName string) []Transaction {
	text = strings.ReplaceAll(text, "\r\n", "\n")

	// Split by transaction indicator.
	parts := strings.Split(text, "ΕΝΔΕΙΞΗ:")

	var result []Transaction

	for i := 1; i < len(parts); i++ {
		block := "ΕΝΔΕΙΞΗ:" + parts[i]

		tx, ok := parseBlock(block, fileName)
		if ok {
			result = append(result, tx)
		}
	}

	return result
}

func parseBlock(block string, fileName string) (Transaction, bool) {
	lines := cleanLines(block)
	if len(lines) == 0 {
		return Transaction{}, false
	}

	indicator := extractRegex(block, `ΕΝΔΕΙΞΗ:\s*([A-Z0-9\-]+)`)
	amount := extractRegex(block, `([0-9]+,[0-9]{2})\s+EUR`)
	card := extractRegex(block, `(441029x+x+\d{4}|441029X+X+\d+)`)
	mcc := extractRegex(block, `(?m)^\s*(\d{4})\s*(GOOGLE-PAY|APPLE-PAY)?\s*$`)
	paymentMethod := extractRegex(block, `(GOOGLE-PAY|APPLE-PAY|MOBW)`)

	if indicator == nil {
		return Transaction{}, false
	}

	merchant := findMerchant(lines)
	description := merchant

	direction := "DEBIT"
	if strings.Contains(block, "PAYROLL") {
		direction = "CREDIT"
	}

	if amount == nil && strings.Contains(block, "PAYROLL") {
		// Payroll amount is harder because PDF extraction is corrupted.
		// Keep raw block and let a later parser version improve this.
		amount = strPtr("")
	}

	if amount == nil {
		return Transaction{}, false
	}

	tx := Transaction{
		AccountID:           0,
		Date:                time.Time{}, // TODO: parse from row prefix when PDF extraction is stabilized
		Indicator:           indicator,
		BankReferenceNumber: indicator,
		MerchantIdentifier:  normalizeMerchant(merchant),
		Description:         description,
		Justification:       description,
		Amount1:             amount,
		Amount:              normalizeAmount(*amount),
		Currency:            "EUR",
		CardMasked:          card,
		MCCCode:             mcc,
		PaymentMethod:       paymentMethod,
		Direction:           direction,
		StatementFileName:   strPtr(fileName),
		RawText:             block,
		ParserVersion:       "piraeus.v1",
		BankCode:            "piraeus",
	}

	return tx, true
}

func cleanLines(s string) []string {
	raw := strings.Split(s, "\n")
	var lines []string

	for _, line := range raw {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}

	return lines
}

func findMerchant(lines []string) *string {
	for _, line := range lines {
		if strings.HasPrefix(line, "ΕΝΔΕΙΞΗ:") {
			continue
		}
		if strings.Contains(line, "EUR") {
			continue
		}
		if strings.Contains(line, "441029") {
			continue
		}
		if regexp.MustCompile(`^\d{4}`).MatchString(line) {
			continue
		}
		if strings.Contains(line, "ΤΡΕΧΟΥΜΕΝΟΣ") {
			continue
		}
		if strings.Contains(line, "Αριθµός") {
			continue
		}

		return strPtr(line)
	}

	return nil
}

func normalizeMerchant(s *string) *string {
	if s == nil {
		return nil
	}

	v := strings.ToLower(*s)
	v = strings.ReplaceAll(v, "_", " ")
	v = regexp.MustCompile(`[^a-z0-9α-ωάέήίόύώϊϋΐΰ]+`).ReplaceAllString(v, "-")
	v = strings.Trim(v, "-")

	return strPtr(v)
}

func normalizeAmount(s string) string {
	return strings.ReplaceAll(s, ",", ".")
}

func extractRegex(s string, pattern string) *string {
	re := regexp.MustCompile(pattern)
	m := re.FindStringSubmatch(s)
	if len(m) < 2 {
		return nil
	}
	v := strings.TrimSpace(m[1])
	return &v
}

func strPtr(s string) *string {
	return &s
}

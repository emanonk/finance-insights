// Package v1 parses Piraeus Bank PDF statements (format version 1).
package v1

import (
	"regexp"
	"strings"

	"github.com/manoskammas/finance-insights/apps/api/internal/parsers"
)

// Compiled regexes for the Piraeus Bank statement layout produced by pdftotext -layout.
var (
	// Account number in page header, e.g. "5009-112563-658"
	reAccount = regexp.MustCompile(`(\d{4}-\d{9}-\d{3})`)

	// Transaction header line:
	//   optional date   ref-indicator  ref-code       description           opt-value-date  amount  balance
	//   08/07/24        2960           EL01P 0442174  ΑΠΚ ΑΓΟΡΑ ΜΕ ΚΑΡΤΑ   08/07/24        71.49   1,987.26ΠΙ
	reTransHeader = regexp.MustCompile(
		`^\s*(\d{2}/\d{2}/\d{2})?\s{1,10}` + // optional date
			`(\d{4})\s+` + // indicator (4-digit code)
			`([A-Z0-9]+\s+[A-Z0-9]+)\s+` + // reference part
			`(.+?)\s{2,}` + // description
			`(?:\d{2}/\d{2}/\d{2}\s+)?` + // optional value date
			`([\d,.]+)\s*` + // debit OR credit amount
			`([\d,.\w]+)$`, // balance
	)

	// Previous-balance summary line — skip
	rePrevBalance = regexp.MustCompile(`Προηγούμενο\s+Υπόλοιπο`)

	// Detail sub-lines inside a transaction
	reEndeiksi   = regexp.MustCompile(`ΕΝΔΕΙΞΗ:\s*(\S+)`)
	reAmount1EUR = regexp.MustCompile(`^[\s\t]*([\d,]+)\s+EUR\s*$`)
	reCardMasked = regexp.MustCompile(`\d{6}[xX]+\d{4}`)
	reMCCLine    = regexp.MustCompile(`^[\s\t]*(\d{4})\s*(GOOGLE-PAY|APPLE-PAY)?\s*$`)
)

// Parser implements parsers.BankParser for Piraeus Bank statement format v1.
type Parser struct{}

// New returns a new Piraeus Bank v1 Parser.
func New() *Parser { return &Parser{} }

func (p *Parser) BankName() string { return "piraeus" }
func (p *Parser) Version() string  { return "v1" }

// Parse extracts all transactions from the PDF at pdfPath.
func (p *Parser) Parse(pdfPath string) ([]parsers.ParsedTransaction, error) {
	raw, err := extractText(pdfPath)
	if err != nil {
		return nil, err
	}

	accountID := ""
	if m := reAccount.FindString(raw); m != "" {
		accountID = m
	}

	var transactions []parsers.ParsedTransaction
	for _, page := range splitPages(raw) {
		transactions = append(transactions, parsePage(page, accountID)...)
	}
	return transactions, nil
}

// parsePage processes one page of pdftotext -layout output.
func parsePage(page, accountID string) []parsers.ParsedTransaction {
	lines := strings.Split(page, "\n")
	var result []parsers.ParsedTransaction

	var current *parsers.ParsedTransaction
	var lastDate string
	inDetail := false

	flush := func() {
		if current != nil {
			if current.AccountID == "" {
				current.AccountID = accountID
			}
			result = append(result, *current)
			current = nil
			inDetail = false
		}
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if rePrevBalance.MatchString(trimmed) {
			continue
		}
		if isPageHeader(trimmed) {
			continue
		}

		if m := reEndeiksi.FindStringSubmatch(line); m != nil {
			if current != nil {
				current.Justification = m[1]
				current.BankReferenceNumber = m[1]
				inDetail = true
			}
			continue
		}

		if inDetail && current != nil {
			if tryParseDetailLine(line, current) {
				continue
			}
		}

		if m := reTransHeader.FindStringSubmatch(line); m != nil {
			flush()

			date := strings.TrimSpace(m[1])
			if date == "" {
				date = lastDate
			} else {
				lastDate = date
			}

			current = &parsers.ParsedTransaction{
				AccountID:               accountID,
				Date:                    normalizeDate(date),
				Indicator:               strings.TrimSpace(m[2]),
				Reference:               strings.TrimSpace(m[3]),
				Description:             strings.TrimSpace(m[4]),
				Direction:               inferDirection(strings.TrimSpace(m[4])),
				Amount:                  normalizeAmount(strings.TrimSpace(m[5])),
				BalanceAfterTransaction: cleanBalance(strings.TrimSpace(m[6])),
			}
			continue
		}

		if inDetail && current != nil && current.MerchantIdentifier == "" {
			candidate := strings.TrimSpace(line)
			if looksLikeMerchant(candidate) {
				current.MerchantIdentifier = candidate
			}
		}
	}
	flush()
	return result
}

// tryParseDetailLine attempts to extract a detail field from a sub-line.
// Returns true if the line was consumed.
func tryParseDetailLine(line string, tx *parsers.ParsedTransaction) bool {
	trimmed := strings.TrimSpace(line)

	if m := reAmount1EUR.FindStringSubmatch(line); m != nil {
		tx.Amount1 = strings.ReplaceAll(m[1], ",", ".")
		return true
	}
	if m := reCardMasked.FindString(line); m != "" {
		tx.CardMasked = m
		return true
	}
	if m := reMCCLine.FindStringSubmatch(line); m != nil {
		tx.MCCCode = m[1]
		if m[2] != "" {
			tx.PaymentMethod = m[2]
		}
		return true
	}
	if tx.MerchantIdentifier == "" && trimmed != "" && looksLikeMerchant(trimmed) {
		tx.MerchantIdentifier = trimmed
		return true
	}
	return false
}

// inferDirection determines "debit" or "credit" based on Greek description keywords.
func inferDirection(description string) string {
	desc := strings.ToUpper(description)
	for _, kw := range []string{"ΜΙΣΘΟΔΟΣΙΑ", "ΜΕΤΑΦΟΡΑ ΣΕ ΛΟΓ", "ΠΙΣΤΩΣΗ"} {
		if strings.Contains(desc, kw) {
			return "credit"
		}
	}
	return "debit"
}

// normalizeDate converts dd/mm/yy → dd/mm/20yy.
func normalizeDate(d string) string {
	if len(d) == 8 && d[2] == '/' && d[5] == '/' {
		return d[:6] + "20" + d[6:]
	}
	return d
}

// normalizeAmount converts European comma-decimal to dot-decimal.
func normalizeAmount(s string) string {
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, ",", ".")
	return s
}

// cleanBalance strips the ΠΙ/ΧΡ suffix and normalises the number.
func cleanBalance(s string) string {
	s = regexp.MustCompile(`[ΠΙΧΡπιχρ]+$`).ReplaceAllString(s, "")
	return normalizeAmount(strings.TrimSpace(s))
}

// looksLikeMerchant returns true if the line looks like a merchant name
// (not a pure number, not a payment-method keyword).
func looksLikeMerchant(s string) bool {
	if s == "" {
		return false
	}
	upper := strings.ToUpper(s)
	if upper == "GOOGLE-PAY" || upper == "APPLE-PAY" {
		return false
	}
	if reAmount1EUR.MatchString(s) || reMCCLine.MatchString(s) || reCardMasked.MatchString(s) {
		return false
	}
	for _, r := range s {
		if r > 127 || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
			return true
		}
	}
	return false
}

// isPageHeader returns true for Piraeus header/footer boilerplate lines.
func isPageHeader(s string) bool {
	for _, kw := range []string{
		"Piraeus", "Κίνηση Λογαριασμού", "Ο Λογαριασμός σας",
		"Στοιχεία Πελάτη", "Αναλυτικά Στοιχεία", "Αριθμός Σελίδας",
		"Ημ/νία", "Κωδ. Αναφοράς", "Αιτιολογία", "Χρέωση", "Πίστωση", "Υπόλοιπο",
		"Νέο Υπόλοιπο", "Αριθμός", "Τύπος", "Νόμισμα",
		"ΤΡΕΧΟΥΜΕΝΟΣ", "ΕΥΡΩ",
		"Από", "Έως",
	} {
		if strings.Contains(s, kw) {
			return true
		}
	}
	return false
}

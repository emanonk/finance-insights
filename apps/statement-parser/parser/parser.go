package parser

import (
	"regexp"
	"strings"
)

// Compiled regexes for the Piraeus Bank statement layout produced by pdftotext -layout.
var (
	// Account number in page header, e.g. "5009-112563-658"
	reAccount = regexp.MustCompile(`(\d{4}-\d{9}-\d{3})`)

	// Transaction header line:
	//   optional date   ref-indicator  ref-code       description           opt-value-date  amount  balance
	//   08/07/24        2960           EL01P 0442174  ΑΠΚ ΑΓΟΡΑ ΜΕ ΚΑΡΤΑ   08/07/24        71.49   1,987.26ΠΙ
	// The reference code block is digits + space + alphanumeric sequence + space + alphanumeric sequence.
	reTransHeader = regexp.MustCompile(
		`^\s*(\d{2}/\d{2}/\d{2})?\s{1,10}` + // optional date
			`(\d{4})\s+` + // indicator (4-digit code)
			`([A-Z0-9]+\s+[A-Z0-9]+)\s+` + // reference part
			`(.+?)\s{2,}` + // description
			`(?:\d{2}/\d{2}/\d{2}\s+)?` + // optional value date
			`([\d,.]+)\s*` + // debit OR credit amount
			`([\d,.\w]+)$`, // balance
	)

	// Same but with credit in a later column (debit column empty).
	// We detect direction by checking which regex group is non-empty.

	// Previous-balance summary line — skip
	rePrevBalance = regexp.MustCompile(`Προηγούμενο\s+Υπόλοιπο`)

	// Detail sub-lines inside a transaction
	reEndeiksi   = regexp.MustCompile(`ΕΝΔΕΙΞΗ:\s*(\S+)`)
	reAmount1EUR = regexp.MustCompile(`^[\s\t]*([\d,]+)\s+EUR\s*$`)
	reCardMasked = regexp.MustCompile(`\d{6}[xX]+\d{4}`)
	reMCCLine    = regexp.MustCompile(`^[\s\t]*(\d{4})\s*(GOOGLE-PAY|APPLE-PAY)?\s*$`)

	// Balance value, e.g. "1,987.26ΠΙ" or "1.987,26ΠΙ"
	reBalance = regexp.MustCompile(`([\d,.']+)\s*ΠΙ\s*$`)

	// Debit/credit amount on a header line — last two numbers before balance
	reAmounts = regexp.MustCompile(`([\d,.]+)\s+([\d,.\w]+)\s*$`)
)

// Parser parses Piraeus Bank PDF statements.
type Parser struct{}

// NewParser returns a new Parser.
func NewParser() *Parser { return &Parser{} }

// Parse extracts all transactions from the PDF at pdfPath.
func (p *Parser) Parse(pdfPath string) ([]Transaction, error) {
	raw, err := extractText(pdfPath)
	if err != nil {
		return nil, err
	}

	accountID := ""
	if m := reAccount.FindString(raw); m != "" {
		accountID = m
	}

	var transactions []Transaction
	pages := splitPages(raw)

	for _, page := range pages {
		txs := parsePage(page, accountID)
		transactions = append(transactions, txs...)
	}

	return transactions, nil
}

// parsePage processes one page of pdftotext -layout output.
func parsePage(page, accountID string) []Transaction {
	lines := strings.Split(page, "\n")
	var result []Transaction

	var current *Transaction
	var lastDate string
	inDetail := false // true once we've seen ΕΝΔΕΙΞΗ for the current transaction

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
		// Skip blank lines and header/footer noise
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

		// Detect ΕΝΔΕΙΞΗ — signals start of detail block for current transaction
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

		// Try to match a transaction header line
		if m := reTransHeader.FindStringSubmatch(line); m != nil {
			flush()

			date := strings.TrimSpace(m[1])
			if date == "" {
				date = lastDate
			} else {
				lastDate = date
			}

			indicator := strings.TrimSpace(m[2])
			reference := strings.TrimSpace(m[3])
			description := strings.TrimSpace(m[4])
			amountRaw := strings.TrimSpace(m[5])
			balanceRaw := strings.TrimSpace(m[6])

			direction := inferDirection(description, amountRaw, balanceRaw)

			current = &Transaction{
				AccountID:               accountID,
				Date:                    normalizeDate(date),
				Indicator:               indicator,
				Reference:               reference,
				Description:             description,
				Direction:               direction,
				Amount:                  normalizeAmount(amountRaw),
				BalanceAfterTransaction: cleanBalance(balanceRaw),
			}
			continue
		}

		// If we reach here and have a current transaction in detail mode,
		// this might be a merchant name or other continuation line.
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

// tryParseDetailLine attempts to extract detail fields from a sub-line.
// Returns true if the line was consumed.
func tryParseDetailLine(line string, tx *Transaction) bool {
	trimmed := strings.TrimSpace(line)

	// Amount like "71,49 EUR"
	if m := reAmount1EUR.FindStringSubmatch(line); m != nil {
		tx.Amount1 = strings.ReplaceAll(m[1], ",", ".")
		return true
	}

	// Masked card number
	if m := reCardMasked.FindString(line); m != "" {
		tx.CardMasked = m
		return true
	}

	// MCC code line: "5945" or "5411 GOOGLE-PAY"
	if m := reMCCLine.FindStringSubmatch(line); m != nil {
		tx.MCCCode = m[1]
		if m[2] != "" {
			tx.PaymentMethod = m[2]
		}
		return true
	}

	// Merchant name — only set once (first descriptive line after ΕΝΔΕΙΞΗ)
	if tx.MerchantIdentifier == "" && trimmed != "" && looksLikeMerchant(trimmed) {
		tx.MerchantIdentifier = trimmed
		return true
	}

	return false
}

// inferDirection determines Debit or Credit based on Greek description keywords.
func inferDirection(description, _, _ string) string {
	desc := strings.ToUpper(description)
	// Credit indicators
	for _, kw := range []string{"ΜΙΣΘΟΔΟΣΙΑ", "ΜΕΤΑΦΟΡΑ ΣΕ ΛΟΓ", "ΠΙΣΤΩΣΗ"} {
		if strings.Contains(desc, kw) {
			return "Credit"
		}
	}
	return "Debit"
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
	// Remove thousands separator dots, replace decimal comma with dot
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, ",", ".")
	return s
}

// cleanBalance strips the ΠΙ/ΧΡ suffix and normalises the number.
func cleanBalance(s string) string {
	s = regexp.MustCompile(`[ΠΙΧΡπιχρ]+$`).ReplaceAllString(s, "")
	return normalizeAmount(strings.TrimSpace(s))
}

// looksLikeMerchant returns true if the trimmed line looks like a merchant name
// (not a number-only line, not a payment method keyword).
func looksLikeMerchant(s string) bool {
	if s == "" {
		return false
	}
	upper := strings.ToUpper(s)
	if upper == "GOOGLE-PAY" || upper == "APPLE-PAY" {
		return false
	}
	// Reject pure-numeric or amount lines
	if reAmount1EUR.MatchString(s) {
		return false
	}
	if reMCCLine.MatchString(s) {
		return false
	}
	if reCardMasked.MatchString(s) {
		return false
	}
	// Must contain at least one letter
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

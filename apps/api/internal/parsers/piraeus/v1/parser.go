// Package v1 parses Piraeus Bank PDF statements (format version 1).
//
// Text is extracted via pdftotext -layout (poppler). The output uses fixed-width
// columns, so debit vs. credit direction is inferred from the column gap between
// the amount and the running balance.
package v1

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/manoskammas/finance-insights/apps/api/internal/parsers"
)

// ── Constants ─────────────────────────────────────────────────────────────────

// debitGapThreshold is the minimum rune distance between the amount and balance
// columns that indicates the amount is in the debit (left) column.
const debitGapThreshold = 20

// ── Regexes ───────────────────────────────────────────────────────────────────

var (
	// Account number in page header, e.g. "5009-112563-658"
	reAccountNumber = regexp.MustCompile(`\d{4}-\d{6}-\d{3}`)

	// Transaction line: optional date, 4-digit indicator, reference, description.
	// The tail (amounts + balance) is parsed separately by extractAmountsAndDirection.
	reTransactionLine = regexp.MustCompile(
		`^(\d{2}/\d{2}/\d{2})?\s+` + // [1] optional date
			`(\d{4})\s+` + // [2] indicator
			`([A-Z0-9]+\s+[A-Z0-9]+)\s+` + // [3] reference  e.g. "EL01P 0442174"
			`(.+?)\s{2,}`, // [4] description (non-greedy, ends at 2+ spaces)
	)

	// isBalanceSummaryLine helpers:
	//   reBalanceSuffix  — line ends with digits.digits + non-ASCII (ΠΙ / ΧΡ or garbled)
	//   reIndicatorRef   — line contains an indicator+reference pair → it's a transaction, not a summary
	reBalanceSuffix = regexp.MustCompile(`\d+\.\d+[^\x00-\x7F\s]+\s*$`)
	reIndicatorRef  = regexp.MustCompile(`\s\d{4}\s+[A-Z][A-Z0-9]*\s+[A-Z0-9]+\s`)

	// Encoding-agnostic ΕΝΔΕΙΞΗ detection:
	//   matches any word that contains at least one non-ASCII rune followed by ":"
	reJustification = regexp.MustCompile(`^\s+\S*[^\x00-\x7F]\S*:\s+(.+?)\s*$`)

	// Detail sub-line patterns
	reAmountEUR  = regexp.MustCompile(`^\s*([\d,]+)\s+EUR\s*$`)
	reCardMasked = regexp.MustCompile(`\d{6}[xX]+\d{4}`)
	reMCCLine    = regexp.MustCompile(`^\s*(\d{4})\s*(GOOGLE-PAY|APPLE-PAY)?\s*$`)

	// Page header / footer boilerplate (both correct UTF-8 and garbled Latin-1)
	rePageBoilerplate = buildBoilerplatePattern()
)

func buildBoilerplatePattern() *regexp.Regexp {
	keywords := []string{
		"Piraeus",
		// Correct UTF-8
		"Κίνηση", "Λογαριασμού", "Ο Λογαριασμός", "Στοιχεία Πελάτη",
		"Αναλυτικά Στοιχεία", "Αριθμός Σελίδας", "Αριθµός Σελίδας",
		"Ημ/νία", "Κωδ. Αναφοράς", "Αιτιολογία",
		"Χρέωση", "Πίστωση", "Υπόλοιπο", "Νέο Υπόλοιπο",
		"Αριθμός", "Τύπος", "Νόμισμα", "ΤΡΕΧΟΥΜΕΝΟΣ", "ΕΥΡΩ", "Από", "Έως",
		// Garbled Latin-1 equivalents
		"Êßíçóç", "Ëïãáñéáóìïý", "Ï Ëïãáñéáóìüò",
		"Óôïé÷åßá ÐåëÜôç", "ÁíáëõôéêÜ Óôïé÷åßá",
		"Áñéèìüò", "Çì/íßá", "Êùä. ÁíáöïñÜò", "Áéôéïëïãßá",
		"×ñÝùóç", "Ðßóôùóç", "Õðüëïéðï", "ÍÝï Õðüëïéðï",
		"Ôýðïò", "Íüìéóìá", "Áðü",
	}
	parts := make([]string, len(keywords))
	for i, kw := range keywords {
		parts[i] = regexp.QuoteMeta(kw)
	}
	return regexp.MustCompile(strings.Join(parts, "|"))
}

// ── Parser ────────────────────────────────────────────────────────────────────

// Parser implements parsers.BankParser for Piraeus Bank statement format v1.
type Parser struct{}

func New() *Parser { return &Parser{} }

func (p *Parser) BankName() string { return "piraeus" }
func (p *Parser) Version() string  { return "v1" }

// Parse extracts all transactions from the PDF at pdfPath.
func (p *Parser) Parse(pdfPath string) ([]parsers.ParsedTransaction, error) {
	fmt.Printf("\n[piraeus/v1] extracting: %s\n", pdfPath)

	out, err := exec.Command("pdftotext", "-layout", "-enc", "UTF-8", pdfPath, "-").Output()
	if err != nil {
		return nil, fmt.Errorf("pdftotext failed — install poppler-utils: %w", err)
	}
	raw := string(out)
	fmt.Printf("[piraeus/v1] extracted ------\n %d bytes\n", len(raw))

	accountNumber := ""
	if m := reAccountNumber.FindString(raw); m != "" {
		accountNumber = m
	}
	fmt.Printf("[piraeus/v1] account: %s\n", accountNumber)

	pages := strings.Split(raw, "\f")
	fmt.Printf("[piraeus/v1] pages: %d\n\n", len(pages))

	var all []parsers.ParsedTransaction
	var lastDate string
	for i, pageText := range pages {
		txs := parsePage(pageText, accountNumber, i+1, &lastDate)
		all = append(all, txs...)
	}

	fmt.Printf("\n[piraeus/v1] total transactions: %d\n\n", len(all))
	return all, nil
}

// ── Page parser ───────────────────────────────────────────────────────────────

type pageState int

const (
	stateHeader pageState = iota // scanning for the previous-balance summary line
	stateBody                    // parsing transaction lines
)

func parsePage(text, accountNumber string, pageNum int, lastDate *string) []parsers.ParsedTransaction {
	lines := strings.Split(text, "\n")
	fmt.Printf("[page %d] %d lines\n", pageNum, len(lines))

	state := stateHeader
	var result []parsers.ParsedTransaction
	var current *parsers.ParsedTransaction

	flush := func() {
		if current == nil {
			return
		}
		if current.AccountID == "" {
			current.AccountID = accountNumber
		}
		printTransaction(pageNum, current)
		result = append(result, *current)
		current = nil
	}

	for li, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// ── State: stateHeader ────────────────────────────────────────────────
		if state == stateHeader {
			if isBalanceSummaryLine(line) {
				printHeader(pageNum, li, line)
				state = stateBody
			}
			continue
		}

		// ── State: stateBody ──────────────────────────────────────────────────

		// Skip page boilerplate (header/footer labels)
		if rePageBoilerplate.MatchString(trimmed) {
			continue
		}

		// ΕΝΔΕΙΞΗ (justification code) — encoding-agnostic
		if m := reJustification.FindStringSubmatch(line); m != nil {
			if current != nil {
				current.Justification = m[1]
				current.BankReferenceNumber = m[1]
				fmt.Printf("[page %d] line[%d] justification: %q\n", pageNum, li, m[1])
			}
			continue
		}

		// Detail sub-lines (only when we have an open transaction)
		if current != nil {
			if consumed, field, val := parseDetailLine(line, current); consumed {
				fmt.Printf("[page %d] line[%d] detail %s=%q\n", pageNum, li, field, val)
				continue
			}
		}

		// Transaction line
		if m := reTransactionLine.FindStringSubmatch(line); m != nil {
			flush()

			date := strings.TrimSpace(m[1])
			if date == "" {
				date = *lastDate
			} else {
				*lastDate = date
			}

			tailStart := len(m[0]) // everything after the matched prefix
			amount, balance, dir := extractAmountsAndDirection(line, tailStart)

			current = &parsers.ParsedTransaction{
				AccountID:               accountNumber,
				Date:                    expandYear(date),
				Indicator:               strings.TrimSpace(m[2]),
				Reference:               strings.TrimSpace(m[3]),
				Description:             strings.TrimSpace(m[4]),
				Direction:               dir,
				Amount:                  amount,
				BalanceAfterTransaction: balance,
			}
			fmt.Printf("[page %d] line[%d] transaction: date=%q indicator=%q ref=%q desc=%q dir=%q amount=%q balance=%q\n",
				pageNum, li,
				current.Date, current.Indicator, current.Reference,
				current.Description, current.Direction, current.Amount, current.BalanceAfterTransaction)
			continue
		}

		// Fallback: unrecognised line after a transaction → possible merchant name
		if current != nil && current.MerchantIdentifier == "" && looksLikeMerchant(trimmed) {
			current.MerchantIdentifier = trimmed
			fmt.Printf("[page %d] line[%d] merchant fallback: %q\n", pageNum, li, trimmed)
		}
	}

	flush()
	return result
}

// ── Detail line parser ────────────────────────────────────────────────────────

func parseDetailLine(line string, tx *parsers.ParsedTransaction) (consumed bool, field, value string) {
	trimmed := strings.TrimSpace(line)

	if m := reAmountEUR.FindStringSubmatch(line); m != nil {
		tx.Amount1 = euroToDecimal(m[1])
		return true, "amountEUR", tx.Amount1
	}
	if m := reCardMasked.FindString(line); m != "" {
		tx.CardMasked = m
		return true, "card", m
	}
	if m := reMCCLine.FindStringSubmatch(line); m != nil {
		tx.MCCCode = m[1]
		if m[2] != "" {
			tx.PaymentMethod = m[2]
			return true, "mcc+pay", m[1] + "/" + m[2]
		}
		return true, "mcc", m[1]
	}
	if tx.MerchantIdentifier == "" && trimmed != "" && looksLikeMerchant(trimmed) {
		tx.MerchantIdentifier = trimmed
		return true, "merchant", trimmed
	}
	return false, "", ""
}

// ── Amount + direction extraction ────────────────────────────────────────────

// extractAmountsAndDirection walks the tail of a transaction line backwards to
// locate the balance and amount fields, then uses the column gap between them
// to infer direction:
//
//	gap ≥ debitGapThreshold  → debit  (amount is in the left/debit column)
//	gap < debitGapThreshold  → credit (amount is in the right/credit column)
func extractAmountsAndDirection(line string, tailStart int) (amount, balance, direction string) {
	tail := line[tailStart:]
	runes := []rune(tail)
	n := len(runes)
	i := n - 1

	// Skip trailing non-digit suffix (e.g. "ÐÉ", "ΠΙ", "ΧΡ")
	for i >= 0 && !isDigit(runes[i]) {
		i--
	}

	// Read balance (digits, dots, commas)
	balEnd := i + 1
	for i >= 0 && isBalanceRune(runes[i]) {
		i--
	}
	rawBalance := string(runes[i+1 : balEnd])

	// Count gap (spaces between balance and amount)
	gapEnd := i
	for i >= 0 && runes[i] == ' ' {
		i--
	}
	gap := gapEnd - i // number of space runes

	// Optional value-date dd/mm/yy — skip it
	if i >= 7 {
		candidate := string(runes[i-7 : i+1])
		if len(candidate) == 8 && candidate[2] == '/' && candidate[5] == '/' {
			i -= 8
			for i >= 0 && runes[i] == ' ' {
				i--
			}
			// recount gap after skipping value date
			gap = gapEnd - i
		}
	}

	// Read amount
	amtEnd := i + 1
	for i >= 0 && isBalanceRune(runes[i]) {
		i--
	}
	rawAmount := string(runes[i+1 : amtEnd])

	if gap >= debitGapThreshold {
		direction = "debit"
	} else {
		direction = "credit"
	}

	return stripThousands(rawAmount), stripThousands(rawBalance), direction
}

func isDigit(r rune) bool       { return r >= '0' && r <= '9' }
func isBalanceRune(r rune) bool { return isDigit(r) || r == '.' || r == ',' }

// ── Line classifiers ──────────────────────────────────────────────────────────

// isBalanceSummaryLine returns true for the "Previous Balance" summary line that
// marks the end of the page header and the start of transaction rows.
// It matches lines that end with a balance+suffix token but are NOT transaction lines.
func isBalanceSummaryLine(line string) bool {
	return reBalanceSuffix.MatchString(line) && !reIndicatorRef.MatchString(line)
}

// looksLikeMerchant returns true if the string contains at least one letter or
// non-ASCII rune, and is not a known sub-line pattern (card, MCC, EUR amount).
func looksLikeMerchant(s string) bool {
	if s == "" {
		return false
	}
	upper := strings.ToUpper(s)
	if upper == "GOOGLE-PAY" || upper == "APPLE-PAY" {
		return false
	}
	if reAmountEUR.MatchString(s) || reMCCLine.MatchString(s) || reCardMasked.MatchString(s) {
		return false
	}
	for _, r := range s {
		if r > 127 || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
			return true
		}
	}
	return false
}

// ── Number helpers ────────────────────────────────────────────────────────────

// stripThousands converts European-formatted numbers (1.234,56 or 1,234.56)
// to plain decimal strings (1234.56).
func stripThousands(s string) string {
	// Determine whether comma or dot is the decimal separator by finding the
	// last separator and checking how many digits follow it.
	lastDot := strings.LastIndex(s, ".")
	lastComma := strings.LastIndex(s, ",")

	var decimal string
	if lastComma > lastDot {
		// comma is decimal separator: "1.234,56"
		s = strings.ReplaceAll(s, ".", "")
		decimal = strings.ReplaceAll(s, ",", ".")
	} else {
		// dot is decimal separator: "1,234.56"
		s = strings.ReplaceAll(s, ",", "")
		decimal = s
	}
	return decimal
}

// euroToDecimal converts comma-decimal amounts found in detail lines ("71,49" → "71.49").
func euroToDecimal(s string) string {
	return strings.ReplaceAll(s, ",", ".")
}

// expandYear converts dd/mm/yy → dd/mm/20yy; passes dd/mm/yyyy unchanged.
func expandYear(d string) string {
	if len(d) == 8 && d[2] == '/' && d[5] == '/' {
		return d[:6] + "20" + d[6:]
	}
	return d
}

// ── Debug output ──────────────────────────────────────────────────────────────

func printHeader(pageNum, lineIndex int, line string) {
	fmt.Printf("[page %d] ── header identified at line[%d]: %q\n", pageNum, lineIndex, strings.TrimSpace(line))
}

func printTransaction(pageNum int, tx *parsers.ParsedTransaction) {
	w := utf8.RuneCountInString
	fmt.Printf("[page %d] ── transaction flushed\n", pageNum)
	fmt.Printf("           date:        %s\n", tx.Date)
	fmt.Printf("           indicator:   %s\n", tx.Indicator)
	fmt.Printf("           reference:   %s\n", tx.Reference)
	fmt.Printf("           description: %s (%d chars)\n", tx.Description, w(tx.Description))
	fmt.Printf("           direction:   %s\n", tx.Direction)
	fmt.Printf("           amount:      %s\n", tx.Amount)
	fmt.Printf("           balance:     %s\n", tx.BalanceAfterTransaction)
	if tx.Justification != "" {
		fmt.Printf("           justif:      %s\n", tx.Justification)
	}
	if tx.MerchantIdentifier != "" {
		fmt.Printf("           merchant:    %s\n", tx.MerchantIdentifier)
	}
	if tx.CardMasked != "" {
		fmt.Printf("           card:        %s\n", tx.CardMasked)
	}
	if tx.MCCCode != "" {
		fmt.Printf("           mcc:         %s  pay=%s\n", tx.MCCCode, tx.PaymentMethod)
	}
}

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
	Indicator               string    `json:"indicator"`
	TransactionType         string    `json:"transactionType"`
	MerchantIdentifier      string    `json:"merchantIdentifier"`
	Amount                  string    `json:"amount"`
	Currency                string    `json:"currency"`
	Direction               string    `json:"direction"`
	MCCCode                 *string   `json:"mccCode,omitempty"`
	CardMasked              *string   `json:"cardMasked,omitempty"`
	PaymentMethod           *string   `json:"paymentMethod,omitempty"`
	BalanceAfterTransaction *string   `json:"balanceAfterTransaction,omitempty"`
	StatementFileName       string    `json:"statementFileName"`
	RawText                 string    `json:"rawText"`
	NeedsReview             bool      `json:"needsReview"`
}

type RawTransactionFrame struct {
	DateLineIndex int
	StartLine     int
	EndLine       int
	Date          time.Time
	Indicator     string
	Lines         []string
	RawText       string
}

type Money struct {
	Amount  string
	Balance string
}

func main() {
	if len(os.Args) < 2 {
		panic("usage: go run main.go statement.pdf")
	}

	fileName := os.Args[1]

	text, err := extractText(fileName)
	fmt.Printf("***************************Extracted text:\n%s\n****************************************", text)
	if err != nil {
		panic(err)
	}

	frames := frameTransactions(text)
	transactions := mapFramesToTransactions(frames, fileName)

	out, err := json.MarshalIndent(transactions, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(out))
}

func extractText(fileName string) (string, error) {
	cmd := exec.Command("pdftotext", "-layout", fileName, "-")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("pdftotext failed: %w", err)
	}

	return string(out), nil
}

// -----------------------------
// STEP 1: RAW TRANSACTION FRAMES
// -----------------------------

func frameTransactions(text string) []RawTransactionFrame {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	lines := cleanLines(text)

	var frames []RawTransactionFrame

	var currentDate time.Time
	var currentDateLineIndex int

	currentStart := -1

	for i, line := range lines {
		if d, ok := extractDateFromLine(line); ok {
			currentDate = d
			currentDateLineIndex = i
		}

		if strings.Contains(line, "ΕΝΔΕΙΞΗ:") {
			if currentStart >= 0 {
				frame := buildFrame(lines, currentStart, i-1, currentDate, currentDateLineIndex)
				frames = append(frames, frame)
			}

			currentStart = i
		}
	}

	if currentStart >= 0 {
		frame := buildFrame(lines, currentStart, len(lines)-1, currentDate, currentDateLineIndex)
		frames = append(frames, frame)
	}

	return frames
}

func buildFrame(lines []string, start int, end int, date time.Time, dateLineIndex int) RawTransactionFrame {
	frameLines := lines[start : end+1]
	raw := strings.Join(frameLines, "\n")

	return RawTransactionFrame{
		DateLineIndex: dateLineIndex,
		StartLine:     start,
		EndLine:       end,
		Date:          date,
		Indicator:     extractIndicator(raw),
		Lines:         frameLines,
		RawText:       raw,
	}
}

func cleanLines(text string) []string {
	var result []string

	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		result = append(result, line)
	}

	return result
}

// -----------------------------
// STEP 2: CLASSIFY + MAP
// -----------------------------

func mapFramesToTransactions(frames []RawTransactionFrame, fileName string) []Transaction {
	var transactions []Transaction

	for _, frame := range frames {

		fmt.Printf(`
			==================================================
			RAW TRANSACTION FRAME
			==================================================

			DateLineIndex : %d
			StartLine     : %d
			EndLine       : %d
			Date          : %s
			Indicator     : %s

			-------------------- LINES -----------------------

			%s

			------------------- RAW TEXT ---------------------

			%s

			======================================================================================================================================================
			`,
			frame.DateLineIndex,
			frame.StartLine,
			frame.EndLine,
			frame.Date.Format("02/01/2006"),
			frame.Indicator,
			strings.Join(frame.Lines, "\n"),
			frame.RawText,
		)

		tx := mapFrameToTransaction(frame, fileName)
		transactions = append(transactions, tx)
	}

	return transactions
}

func mapFrameToTransaction(frame RawTransactionFrame, fileName string) Transaction {
	transactionType := classifyTransaction(frame)
	direction := detectDirection(frame)

	metadata := parseMetadataByType(frame, transactionType)

	needsReview := false
	if metadata.Amount == "" {
		needsReview = true
	}

	return Transaction{
		AccountID:               0,
		Date:                    frame.Date,
		Indicator:               frame.Indicator,
		TransactionType:         transactionType,
		MerchantIdentifier:      metadata.MerchantIdentifier,
		Amount:                  metadata.Amount,
		Currency:                "EUR",
		Direction:               direction,
		MCCCode:                 metadata.MCCCode,
		CardMasked:              metadata.CardMasked,
		PaymentMethod:           metadata.PaymentMethod,
		BalanceAfterTransaction: metadata.Balance,
		StatementFileName:       fileName,
		RawText:                 frame.RawText,
		NeedsReview:             needsReview,
	}
}

type ParsedMetadata struct {
	MerchantIdentifier string
	Amount             string
	Balance            *string
	MCCCode            *string
	CardMasked         *string
	PaymentMethod      *string
}

func classifyTransaction(frame RawTransactionFrame) string {
	raw := strings.ToUpper(frame.RawText)
	indicator := strings.ToUpper(frame.Indicator)

	switch {
	case strings.Contains(indicator, "PAYROLL SYSTEM"):
		return "PAYROLL"
	case strings.HasPrefix(indicator, "PO"):
		return "CARD_PAYMENT"
	case strings.HasPrefix(indicator, "PX"):
		return "BILL_PAYMENT"
	case strings.HasPrefix(indicator, "AT"):
		return "ATM_OR_BRANCH"
	case strings.HasPrefix(indicator, "MB"):
		return "MOBILE_BANKING"
	case strings.HasPrefix(indicator, "F"):
		return "TRANSFER"
	case strings.Contains(raw, "ΠΡΟΜΗΘΕΙΑ"):
		return "FEE"
	default:
		return "UNKNOWN"
	}
}

func detectDirection(frame RawTransactionFrame) string {
	raw := strings.ToUpper(frame.RawText)

	if strings.Contains(raw, "PAYROLL") ||
		strings.Contains(raw, "ΜΙΣΘΟΔΟΣΙΑ") ||
		strings.Contains(raw, "ΠΙΣΤΩΣ") {
		return "CREDIT"
	}

	return "DEBIT"
}

// -----------------------------
// STEP 3: TYPE-SPECIFIC PARSERS
// -----------------------------

func parseMetadataByType(frame RawTransactionFrame, transactionType string) ParsedMetadata {
	switch transactionType {
	case "CARD_PAYMENT":
		return parseCardPayment(frame)
	case "BILL_PAYMENT":
		return parseBillPayment(frame)
	case "PAYROLL":
		return parsePayroll(frame)
	case "ATM_OR_BRANCH":
		return parseGenericBanking(frame)
	case "MOBILE_BANKING":
		return parseGenericBanking(frame)
	case "TRANSFER":
		return parseGenericBanking(frame)
	default:
		return parseGenericBanking(frame)
	}
}

func parseCardPayment(frame RawTransactionFrame) ParsedMetadata {
	raw := frame.RawText

	amount := extractString(raw, `([0-9]+,[0-9]{2})\s+EUR`)
	money := extractMoneyFromIndicatorLine(frame.Lines[0])

	if amount == "" {
		amount = money.Amount
	}

	return ParsedMetadata{
		MerchantIdentifier: extractMerchantIdentifier(frame.Lines),
		Amount:             normalizeAmount(amount),
		Balance:            strPtrIfNotEmpty(normalizeAmount(money.Balance)),
		MCCCode:            extractPtr(raw, `(?m)^\s*([0-9]{4})\s*(GOOGLE-PAY|APPLE-PAY)?\s*$`),
		CardMasked:         extractPtr(raw, `(441029[xX]+[0-9]+|441029X+[0-9]+)`),
		PaymentMethod:      extractPtr(raw, `(GOOGLE-PAY|APPLE-PAY|APPLE PAY|GOOGLE PAY)`),
	}
}

func parseBillPayment(frame RawTransactionFrame) ParsedMetadata {
	money := extractMoneyFromIndicatorLine(frame.Lines[0])

	return ParsedMetadata{
		MerchantIdentifier: extractMerchantIdentifier(frame.Lines),
		Amount:             normalizeAmount(money.Amount),
		Balance:            strPtrIfNotEmpty(normalizeAmount(money.Balance)),
		PaymentMethod:      extractPtr(frame.RawText, `(MOBW)`),
	}
}

func parsePayroll(frame RawTransactionFrame) ParsedMetadata {
	money := extractMoneyFromIndicatorLine(frame.Lines[0])

	return ParsedMetadata{
		MerchantIdentifier: extractMerchantIdentifier(frame.Lines),
		Amount:             normalizeAmount(money.Amount),
		Balance:            strPtrIfNotEmpty(normalizeAmount(money.Balance)),
	}
}

func parseGenericBanking(frame RawTransactionFrame) ParsedMetadata {
	money := extractMoneyFromIndicatorLine(frame.Lines[0])

	return ParsedMetadata{
		MerchantIdentifier: extractMerchantIdentifier(frame.Lines),
		Amount:             normalizeAmount(money.Amount),
		Balance:            strPtrIfNotEmpty(normalizeAmount(money.Balance)),
		PaymentMethod:      extractPtr(frame.RawText, `(MOBW|WEB|ATM|BRANCH)`),
	}
}

// -----------------------------
// MONEY EXTRACTION
// -----------------------------

func extractMoneyFromIndicatorLine(line string) Money {
	beforeIndicator := strings.Split(line, "ΕΝΔΕΙΞΗ:")[0]

	// Important:
	// Do not over-normalize corrupted PDF chars into digits.
	// First try real readable money patterns.
	matches := findReadableMoneyTokens(beforeIndicator)

	if len(matches) >= 2 {
		return Money{
			Amount:  matches[len(matches)-2],
			Balance: matches[len(matches)-1],
		}
	}

	if len(matches) == 1 {
		return Money{
			Amount: matches[0],
		}
	}

	// Last fallback: use raw noisy numbers.
	// Keep empty if not confident.
	return Money{}
}

func findReadableMoneyTokens(s string) []string {
	// Matches:
	// 71,49
	// 2.425,45
	// 2425,45
	// 400,00
	re := regexp.MustCompile(`[0-9]{1,3}(?:\.[0-9]{3})*,[0-9]{2}|[0-9]{1,6},[0-9]{2}`)
	return re.FindAllString(s, -1)
}

func normalizeAmount(amount string) string {
	amount = strings.TrimSpace(amount)
	if amount == "" {
		return ""
	}

	amount = strings.ReplaceAll(amount, ".", "")
	amount = strings.ReplaceAll(amount, ",", ".")

	return amount
}

// -----------------------------
// FIELD EXTRACTORS
// -----------------------------

func extractDateFromLine(line string) (time.Time, bool) {
	re := regexp.MustCompile(`\b([0-9]{2})/([0-9]{2})/([0-9]{2,4})\b`)
	m := re.FindStringSubmatch(line)
	if len(m) == 0 {
		return time.Time{}, false
	}

	dateText := m[1] + "/" + m[2] + "/" + m[3]

	layout := "02/01/2006"
	if len(m[3]) == 2 {
		layout = "02/01/06"
	}

	d, err := time.Parse(layout, dateText)
	if err != nil {
		return time.Time{}, false
	}

	return d, true
}

func extractIndicator(raw string) string {
	re := regexp.MustCompile(`ΕΝΔΕΙΞΗ:\s*([^\n\r]+)`)
	m := re.FindStringSubmatch(raw)
	if len(m) < 2 {
		return ""
	}

	return strings.TrimSpace(m[1])
}

func extractMerchantIdentifier(lines []string) string {
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}
		if strings.Contains(line, "ΕΝΔΕΙΞΗ:") {
			continue
		}
		if strings.Contains(line, "EUR") {
			continue
		}
		if strings.Contains(line, "441029") {
			continue
		}
		if regexp.MustCompile(`^\d{4}(\s|$)`).MatchString(line) {
			continue
		}
		if strings.Contains(line, "GOOGLE-PAY") || strings.Contains(line, "APPLE-PAY") {
			continue
		}
		if looksLikeFooter(line) {
			continue
		}

		return normalizeSpaces(line)
	}

	return ""
}

func extractString(text string, pattern string) string {
	re := regexp.MustCompile(pattern)
	m := re.FindStringSubmatch(text)
	if len(m) < 2 {
		return ""
	}

	return strings.TrimSpace(m[1])
}

func extractPtr(text string, pattern string) *string {
	v := extractString(text, pattern)
	if v == "" {
		return nil
	}

	return &v
}

func normalizeSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func strPtrIfNotEmpty(s string) *string {
	if strings.TrimSpace(s) == "" {
		return nil
	}

	return &s
}

func looksLikeFooter(line string) bool {
	footerTokens := []string{
		"ΤΡΕΧΟΥΜΕΝΟΣ",
		"ΕΥΡΩ",
		"ΕΥΡΩ",
		"Αριθμός",
		"Αριθµός",
		"Σελίδας",
		"ΚΑΜΜΑΣ",
		"ΕΠΙΔΑΥΡΟΥ",
		"5009-",
		"2009-",
		"Από",
		"Έως",
		"Εως",
	}

	for _, token := range footerTokens {
		if strings.Contains(line, token) {
			return true
		}
	}

	return false
}

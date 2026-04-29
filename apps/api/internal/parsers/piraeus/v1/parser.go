// Package v1 parses Piraeus Bank PDF statements (format version 1).
//
// Text is extracted via pdftotext -layout (poppler). Pages are split on form-feed;
// lines are collected into raw transaction groups which are then parsed and
// post-processed to derive direction and amount from running balance changes.
package v1

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/manoskammas/finance-insights/apps/api/internal/parsers"
)

// ── Parser ────────────────────────────────────────────────────────────────────

// Parser implements parsers.BankParser for Piraeus Bank statement format v1.
type Parser struct{}

func New() *Parser { return &Parser{} }

func (p *Parser) BankName() string { return "piraeus" }
func (p *Parser) Version() string  { return "v1" }

// Parse extracts all transactions from the PDF at pdfPath.
func (p *Parser) Parse(pdfPath string) ([]parsers.ParsedTransaction, error) {
	text, err := extractText(pdfPath)
	if err != nil {
		return nil, fmt.Errorf("pdftotext failed — install poppler-utils: %w", err)
	}

	pages := strings.Split(text, "\f")

	var headerLines []string
	isSetHeader := true
	var rawTransactions [][]string

	reColSplit := regexp.MustCompile(`\s{2,}`)
	reAccountNumber := regexp.MustCompile(`\d{4}-\d{6}-\d{3}`)

	for pageIndex, page := range pages {
		lines := strings.Split(strings.ReplaceAll(page, "\r\n", "\n"), "\n")
		isHeader := false

		var transaction []string

		for i, line := range lines {
			trimmedLine := strings.TrimSpace(line)

			if trimmedLine == "1" {
				isHeader = true
			}

			if isHeader && isSetHeader {
				headerLines = append(headerLines, line)
			}

			if !isHeader {
				cols := reColSplit.Split(strings.TrimSpace(line), -1)

				if len(cols) >= 4 || trimmedLine == "" {
					if len(transaction) > 2 {
						details := "PAGE:" + fmt.Sprint(pageIndex) + " LINE:" + fmt.Sprint(i)
						transaction = append(transaction, details)
						rawTransactions = append(rawTransactions, transaction)
					}
					transaction = nil
				}

				transaction = append(transaction, trimmedLine)
			}

			if strings.Contains(trimmedLine, "Ðñïçãïýìåíï Õðüëïéðï") {
				isHeader = false
				isSetHeader = false
			}
		}
	}

	if len(headerLines) < 13 {
		return nil, fmt.Errorf("could not find account number: header too short (%d lines)", len(headerLines))
	}
	accountNumber := reAccountNumber.FindString(headerLines[12])

	var parsed []parsers.ParsedTransaction
	var previousDate time.Time
	var previousBalanceAfter int

	for i, rawTx := range rawTransactions {
		tx := parseTransaction(rawTx, accountNumber)

		if tx.Date.IsZero() {
			tx.Date = previousDate
		} else {
			previousDate = tx.Date
		}

		if i > 0 {
			tx.BalanceBefore = previousBalanceAfter

			if tx.BalanceAfter < previousBalanceAfter {
				tx.Direction = "debit"
				tx.Amount = previousBalanceAfter - tx.BalanceAfter
			} else {
				tx.Direction = "credit"
				tx.Amount = tx.BalanceAfter - previousBalanceAfter
			}
		}

		previousBalanceAfter = tx.BalanceAfter
		parsed = append(parsed, tx)
	}

	return parsed, nil
}

// ── Transaction parser ────────────────────────────────────────────────────────

func parseTransaction(rawTx []string, accountNumber string) parsers.ParsedTransaction {
	firstRow := rawTx[0]
	cols := regexp.MustCompile(`\s{2,}`).Split(strings.TrimSpace(firstRow), -1)

	var tx parsers.ParsedTransaction
	tx.RawData = rawTx
	tx.AccountID = accountNumber

	if len(cols) > 0 && isDate(cols[0]) {
		tx.Date = parseDate(cols[0])
		cols = cols[1:]
	}

	if len(cols) > 0 {
		ref := cols[0]
		tx.BankReference = &ref
	}

	if len(cols) > 0 {
		tx.BalanceAfter = parseAmount(cols[len(cols)-1])
	}

	for _, row := range rawTx {
		if strings.Contains(row, ":") {
			parts := strings.SplitN(row, ":", 2)
			ref := strings.TrimSpace(parts[1])
			if ref != "" {
				tx.TransactionReference = &ref
				break
			}
		}
	}

	if len(rawTx) > 2 {
		merchant := rawTx[2]
		tx.MerchantIdentifier = &merchant
	}

	return tx
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func extractText(pdfPath string) (string, error) {
	out, err := exec.Command("pdftotext", "-layout", "-enc", "UTF-8", pdfPath, "-").Output()
	return string(out), err
}

func isDate(value string) bool {
	_, err := time.Parse("02/01/06", value)
	return err == nil
}

func parseDate(value string) time.Time {
	t, _ := time.Parse("02/01/06", value)
	return t
}

func parseAmount(value string) int {
	value = strings.TrimSpace(value)
	value = regexp.MustCompile(`[^0-9,.-]`).ReplaceAllString(value, "")
	value = strings.ReplaceAll(value, ".", "")
	value = strings.ReplaceAll(value, ",", "")
	amount, _ := strconv.Atoi(value)
	return amount
}

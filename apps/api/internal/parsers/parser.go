// Package parsers defines the shared interfaces and types for bank statement parsers.
package parsers

import "time"

// ParsedTransaction is the normalized output of any bank statement parser.
type ParsedTransaction struct {
	AccountID            string
	Date                 time.Time
	BankReference        *string
	TransactionReference *string
	MerchantIdentifier   *string
	BalanceBefore        int
	BalanceAfter         int
	Amount               int
	Direction            string // "debit" or "credit"
	RawData              []string
}

// BankParser is implemented by each versioned parser for a specific bank.
type BankParser interface {
	BankName() string
	Version() string
	Parse(pdfPath string) ([]ParsedTransaction, error)
}

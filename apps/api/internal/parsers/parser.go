// Package parsers defines the shared interfaces and types for bank statement parsers.
package parsers

// ParsedTransaction is the raw, unnormalized output of any bank statement parser.
// Fields are strings to preserve the source representation; normalization to domain
// types happens in the service layer.
type ParsedTransaction struct {
	AccountID               string
	Date                    string
	BankReferenceNumber     string
	Justification           string
	Indicator               string
	MerchantIdentifier      string
	Amount1                 string
	MCCCode                 string
	CardMasked              string
	Reference               string
	Description             string
	PaymentMethod           string
	Direction               string
	Amount                  string
	BalanceAfterTransaction string
}

// BankParser is implemented by each versioned parser for a specific bank.
type BankParser interface {
	BankName() string
	Version() string
	Parse(pdfPath string) ([]ParsedTransaction, error)
}

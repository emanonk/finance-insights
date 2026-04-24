package parser

// Transaction holds all fields for a single bank statement entry.
type Transaction struct {
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

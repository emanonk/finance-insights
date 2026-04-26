package domain

import "time"

// Transaction is a normalized financial record parsed from a bank statement.
//
// Amount and BalanceAfterTransaction are strings to preserve exact numeric
// precision across the database boundary (numeric(14,2) column).
type Transaction struct {
	ID                      int64
	AccountID               int64
	Date                    time.Time
	BankReferenceNumber     *string
	Justification           *string
	Indicator               *string
	MerchantIdentifier      *string
	Amount1                 *string
	MCCCode                 *string
	CardMasked              *string
	Reference               *string
	Description             *string
	PaymentMethod           *string
	Direction               string
	Amount                  string
	BalanceAfterTransaction *string
	StatementFileName       *string
}

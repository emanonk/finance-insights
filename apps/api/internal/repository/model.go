// Package repository persists domain entities to PostgreSQL.
package repository

import "time"

// Tag is a row in the tags table.
type Tag struct {
	ID   int64
	Name string
	Type string // "primary" or "secondary"
}

// Merchant is a row in the merchants table with its associated tags preloaded.
type Merchant struct {
	ID             int64
	IdentifierName string
	PrimaryTag     Tag
	SecondaryTags  []Tag
	DefaultTitle   *string
}

// IdentifierCount holds a merchant_identifier value, how many times it appears
// in the transactions table, and the optional Merchant record for it.
type IdentifierCount struct {
	Identifier string
	Count      int
	Merchant   *Merchant
}

// Transaction is a row in the transactions table.
//
// Amount and BalanceAfterTransaction are stored as strings to preserve exact
// numeric precision across the database boundary (the underlying column type
// is numeric(14,2)).
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

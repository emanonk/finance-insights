// Package repository persists domain entities to PostgreSQL.
package repository

import (
	"time"

	"github.com/google/uuid"
)

// Statement is a row in the statements table.
type Statement struct {
	ID         uuid.UUID
	FileName   string
	StoredPath string
	UploadedAt time.Time
}

// Transaction is a row in the transactions table.
//
// Amount and BalanceAfterTransaction are stored as strings to preserve exact
// numeric precision across the database boundary (the underlying column type
// is numeric(14,2)).
type Transaction struct {
	ID                      uuid.UUID
	StatementID             uuid.UUID
	AccountID               *string
	Date                    time.Time
	MerchantIdentifier      *string
	Description             string
	Direction               string
	Amount                  string
	BalanceAfterTransaction *string
	MCCCode                 *string
	CardMasked              *string
	Reference               *string
	BankReferenceNumber     *string
	PaymentMethod           *string
	CreatedAt               time.Time
}

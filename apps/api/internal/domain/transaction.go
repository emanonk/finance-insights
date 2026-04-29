package domain

import "time"

// Transaction is a normalized financial record parsed from a bank statement.
type Transaction struct {
	ID                   int64
	AccountID            int64
	Date                 time.Time
	BankReference        *string
	TransactionReference *string
	MerchantIdentifier   *string
	BalanceBefore        int
	BalanceAfter         int
	Amount               int
	Direction            string
	RawData              []string
	StatementFileName    *string
}

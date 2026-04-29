package service

import (
	"testing"
	"time"

	"github.com/manoskammas/finance-insights/apps/api/internal/parsers"
)

const testAccountID int64 = 1

func ptr(s string) *string { return &s }

func TestToDomainTransactions_ValidRow(t *testing.T) {
	t.Parallel()

	date := time.Date(2024, 7, 8, 0, 0, 0, 0, time.UTC)
	parsed := []parsers.ParsedTransaction{
		{
			Date:               date,
			Direction:          "debit",
			Amount:             7149,
			BalanceBefore:      198726,
			BalanceAfter:       191577,
			MerchantIdentifier: ptr("SOME MERCHANT"),
			BankReference:      ptr("2960 EL01P 0442174"),
		},
	}

	out, err := toDomainTransactions(testAccountID, "statement.pdf", parsed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 row, got %d", len(out))
	}
	got := out[0]
	if got.AccountID != testAccountID {
		t.Errorf("AccountID = %v, want %v", got.AccountID, testAccountID)
	}
	if !got.Date.Equal(date) {
		t.Errorf("Date = %v, want %v", got.Date, date)
	}
	if got.Amount != 7149 {
		t.Errorf("Amount = %d, want 7149", got.Amount)
	}
	if got.BalanceAfter != 191577 {
		t.Errorf("BalanceAfter = %d, want 191577", got.BalanceAfter)
	}
	if got.MerchantIdentifier == nil || *got.MerchantIdentifier != "SOME MERCHANT" {
		t.Errorf("MerchantIdentifier = %v", got.MerchantIdentifier)
	}
	if got.StatementFileName == nil || *got.StatementFileName != "statement.pdf" {
		t.Errorf("StatementFileName = %v", got.StatementFileName)
	}
}

func TestToDomainTransactions_SkipsIncompleteRows(t *testing.T) {
	t.Parallel()

	date := time.Date(2024, 7, 8, 0, 0, 0, 0, time.UTC)
	parsed := []parsers.ParsedTransaction{
		{Date: time.Time{}, Direction: "debit", Amount: 100},   // zero date → skip
		{Date: date, Direction: "", Amount: 100},               // no direction → skip
		{Date: date, Direction: "debit", Amount: 0},            // zero amount → skip
	}
	out, err := toDomainTransactions(testAccountID, "s.pdf", parsed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 0 {
		t.Errorf("expected 0 rows (all incomplete), got %d", len(out))
	}
}

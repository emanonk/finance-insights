package service

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/manoskammas/finance-insights/apps/api/internal/parser"
)

func TestToDomainTransactions_ValidRow(t *testing.T) {
	t.Parallel()

	statementID := uuid.New()
	parsed := []parser.ParsedTransaction{
		{
			AccountID:               "5009-112563-658",
			Date:                    "08/07/2024",
			Description:             "CARD PURCHASE",
			Direction:               "Debit",
			Amount:                  "71.49",
			BalanceAfterTransaction: "1987.26",
			MerchantIdentifier:      "SOME MERCHANT",
			MCCCode:                 "5411",
		},
	}

	out, err := toDomainTransactions(statementID, parsed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 row, got %d", len(out))
	}
	got := out[0]
	if got.StatementID != statementID {
		t.Errorf("StatementID = %v, want %v", got.StatementID, statementID)
	}
	wantDate := time.Date(2024, 7, 8, 0, 0, 0, 0, time.UTC)
	if !got.Date.Equal(wantDate) {
		t.Errorf("Date = %v, want %v", got.Date, wantDate)
	}
	if got.Amount != "71.49" {
		t.Errorf("Amount = %q, want %q", got.Amount, "71.49")
	}
	if got.BalanceAfterTransaction == nil || *got.BalanceAfterTransaction != "1987.26" {
		t.Errorf("Balance = %v, want 1987.26", got.BalanceAfterTransaction)
	}
	if got.MerchantIdentifier == nil || *got.MerchantIdentifier != "SOME MERCHANT" {
		t.Errorf("MerchantIdentifier = %v", got.MerchantIdentifier)
	}
}

func TestToDomainTransactions_SkipsIncompleteRows(t *testing.T) {
	t.Parallel()

	parsed := []parser.ParsedTransaction{
		{Date: "", Description: "no date", Direction: "Debit", Amount: "1"},
		{Date: "08/07/2024", Description: "", Direction: "Debit", Amount: "1"},
		{Date: "08/07/2024", Description: "no amount", Direction: "Debit", Amount: ""},
		{Date: "08/07/2024", Description: "no dir", Direction: "", Amount: "1"},
	}
	out, err := toDomainTransactions(uuid.New(), parsed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 0 {
		t.Errorf("expected 0 rows (all incomplete), got %d", len(out))
	}
}

func TestToDomainTransactions_InvalidAmountFails(t *testing.T) {
	t.Parallel()

	parsed := []parser.ParsedTransaction{
		{Date: "08/07/2024", Description: "bad", Direction: "Debit", Amount: "not-a-number"},
	}
	if _, err := toDomainTransactions(uuid.New(), parsed); err == nil {
		t.Fatal("expected error for invalid amount, got nil")
	}
}

func TestToDomainTransactions_InvalidDateFails(t *testing.T) {
	t.Parallel()

	parsed := []parser.ParsedTransaction{
		{Date: "bad-date", Description: "bad", Direction: "Debit", Amount: "1.00"},
	}
	if _, err := toDomainTransactions(uuid.New(), parsed); err == nil {
		t.Fatal("expected error for invalid date, got nil")
	}
}

func TestParseDate_AcceptsBothLayouts(t *testing.T) {
	t.Parallel()

	cases := []string{"08/07/2024", "2024-07-08"}
	want := time.Date(2024, 7, 8, 0, 0, 0, 0, time.UTC)
	for _, s := range cases {
		got, err := parseDate(s)
		if err != nil {
			t.Errorf("parseDate(%q) error = %v", s, err)
			continue
		}
		if !got.Equal(want) {
			t.Errorf("parseDate(%q) = %v, want %v", s, got, want)
		}
	}
}

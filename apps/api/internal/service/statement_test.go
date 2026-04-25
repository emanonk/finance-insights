package service

import (
	"testing"
	"time"

	"github.com/manoskammas/finance-insights/apps/api/internal/parser"
)

const testAccountID int64 = 1

func TestToDomainTransactions_ValidRow(t *testing.T) {
	t.Parallel()

	parsed := []parser.ParsedTransaction{
		{
			Date:                    "08/07/2024",
			Description:             "CARD PURCHASE",
			Direction:               "Debit",
			Amount:                  "71.49",
			BalanceAfterTransaction: "1987.26",
			MerchantIdentifier:      "SOME MERCHANT",
			MCCCode:                 "5411",
			Justification:           "justification text",
			Indicator:               "D",
			Amount1:                 "71.49",
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
	if got.Justification == nil || *got.Justification != "justification text" {
		t.Errorf("Justification = %v", got.Justification)
	}
	if got.Indicator == nil || *got.Indicator != "D" {
		t.Errorf("Indicator = %v", got.Indicator)
	}
	if got.Amount1 == nil || *got.Amount1 != "71.49" {
		t.Errorf("Amount1 = %v", got.Amount1)
	}
	if got.StatementFileName == nil || *got.StatementFileName != "statement.pdf" {
		t.Errorf("StatementFileName = %v", got.StatementFileName)
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
	out, err := toDomainTransactions(testAccountID, "s.pdf", parsed)
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
	if _, err := toDomainTransactions(testAccountID, "s.pdf", parsed); err == nil {
		t.Fatal("expected error for invalid amount, got nil")
	}
}

func TestToDomainTransactions_InvalidDateFails(t *testing.T) {
	t.Parallel()

	parsed := []parser.ParsedTransaction{
		{Date: "bad-date", Description: "bad", Direction: "Debit", Amount: "1.00"},
	}
	if _, err := toDomainTransactions(testAccountID, "s.pdf", parsed); err == nil {
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

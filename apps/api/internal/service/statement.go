package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/manoskammas/finance-insights/apps/api/internal/domain"
	"github.com/manoskammas/finance-insights/apps/api/internal/parsers"
)

// transactionWriter bulk-persists transactions inside the given tx.
type transactionWriter interface {
	InsertBatch(ctx context.Context, tx pgx.Tx, txs []domain.Transaction) error
}

// statementParser parses a PDF file on disk into a ParseResult.
type statementParser interface {
	Parse(bankName, pdfPath string) (parsers.ParseResult, error)
}

// bankLookup resolves the bank name for a given account id.
type bankLookup interface {
	BankName(ctx context.Context, accountID int64) (string, error)
}

// Statement orchestrates statement ingest: save the upload, parse it,
// and persist the transactions atomically.
type Statement struct {
	pool       *pgxpool.Pool
	parser     statementParser
	txs        transactionWriter
	accounts   bankLookup
	storageDir string
}

// NewStatement constructs a Statement service.
func NewStatement(
	pool *pgxpool.Pool,
	p statementParser,
	txs transactionWriter,
	accounts bankLookup,
	storageDir string,
) *Statement {
	return &Statement{
		pool:       pool,
		parser:     p,
		txs:        txs,
		accounts:   accounts,
		storageDir: storageDir,
	}
}

// IngestResult is returned to clients after a successful upload.
type IngestResult struct {
	FileName         string
	TransactionCount int
}

// Ingest saves the uploaded PDF to disk, parses it with the bank's registered
// parsers, and persists all parsed transactions in a single DB transaction.
// accountID must be a valid FK into the accounts table.
func (s *Statement) Ingest(ctx context.Context, accountID int64, fileName string, r io.Reader) (IngestResult, error) {
	bankName, err := s.accounts.BankName(ctx, accountID)
	if err != nil {
		return IngestResult{}, fmt.Errorf("lookup account: %w", err)
	}

	storedPath, err := s.saveToDisk(fileName, r)
	if err != nil {
		return IngestResult{}, err
	}

	result, err := s.parser.Parse(bankName, storedPath)
	if err != nil {
		_ = os.Remove(storedPath)
		return IngestResult{}, fmt.Errorf("parse statement: %w", err)
	}

	domainTxs, err := toDomainTransactions(accountID, fileName, result.Transactions)
	if err != nil {
		_ = os.Remove(storedPath)
		return IngestResult{}, fmt.Errorf("normalize transactions: %w", err)
	}

	if err := s.persist(ctx, domainTxs); err != nil {
		_ = os.Remove(storedPath)
		return IngestResult{}, err
	}

	return IngestResult{
		FileName:         fileName,
		TransactionCount: len(domainTxs),
	}, nil
}

func (s *Statement) saveToDisk(fileName string, r io.Reader) (string, error) {
	dir := filepath.Join(s.storageDir, "statements")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create storage dir: %w", err)
	}

	safeName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(fileName))
	path := filepath.Join(dir, safeName)
	f, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("create storage file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, r); err != nil {
		_ = os.Remove(path)
		return "", fmt.Errorf("write storage file: %w", err)
	}
	return path, nil
}

func (s *Statement) persist(ctx context.Context, txs []domain.Transaction) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := s.txs.InsertBatch(ctx, tx, txs); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

// toDomainTransactions converts parser output into domain transactions.
// Transactions missing required fields are skipped; malformed values fail fast.
func toDomainTransactions(accountID int64, fileName string, parsed []parsers.ParsedTransaction) ([]domain.Transaction, error) {
	out := make([]domain.Transaction, 0, len(parsed))
	for i, p := range parsed {
		if strings.TrimSpace(p.Date) == "" ||
			strings.TrimSpace(p.Direction) == "" ||
			strings.TrimSpace(p.Amount) == "" ||
			strings.TrimSpace(p.Description) == "" {
			continue
		}

		date, err := parseDate(p.Date)
		if err != nil {
			return nil, fmt.Errorf("transaction #%d date %q: %w", i, p.Date, err)
		}
		amount, err := normalizeAmountString(p.Amount)
		if err != nil {
			return nil, fmt.Errorf("transaction #%d amount %q: %w", i, p.Amount, err)
		}

		desc := p.Description
		t := domain.Transaction{
			AccountID:           accountID,
			Date:                date,
			BankReferenceNumber: nullableString(p.BankReferenceNumber),
			Justification:       nullableString(p.Justification),
			Indicator:           nullableString(p.Indicator),
			MerchantIdentifier:  nullableString(p.MerchantIdentifier),
			MCCCode:             nullableString(p.MCCCode),
			CardMasked:          nullableString(p.CardMasked),
			Reference:           nullableString(p.Reference),
			Description:         &desc,
			PaymentMethod:       nullableString(p.PaymentMethod),
			Direction:           strings.ToLower(p.Direction),
			Amount:              amount,
			StatementFileName:   &fileName,
		}

		if a1, ok, err := optionalAmount(p.Amount1); err != nil {
			return nil, fmt.Errorf("transaction #%d amount1 %q: %w", i, p.Amount1, err)
		} else if ok {
			t.Amount1 = &a1
		}

		if balance, ok, err := optionalAmount(p.BalanceAfterTransaction); err != nil {
			return nil, fmt.Errorf("transaction #%d balance %q: %w", i, p.BalanceAfterTransaction, err)
		} else if ok {
			t.BalanceAfterTransaction = &balance
		}

		out = append(out, t)
	}
	return out, nil
}

// parseDate accepts dd/mm/yyyy (parser output) or ISO YYYY-MM-DD.
func parseDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	for _, layout := range []string{"02/01/2006", "2006-01-02"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC(), nil
		}
	}
	return time.Time{}, errors.New("unrecognized date format")
}

// normalizeAmountString ensures the parser-provided amount is a valid decimal
// expressible as numeric(14,2). It returns the canonical dot-decimal form.
func normalizeAmountString(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", errors.New("empty amount")
	}
	if _, err := strconv.ParseFloat(s, 64); err != nil {
		return "", fmt.Errorf("not a number: %w", err)
	}
	return s, nil
}

func optionalAmount(s string) (string, bool, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", false, nil
	}
	a, err := normalizeAmountString(s)
	if err != nil {
		return "", false, err
	}
	return a, true, nil
}

func nullableString(s string) *string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return &s
}

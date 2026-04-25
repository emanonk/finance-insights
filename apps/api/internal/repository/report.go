package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ReportRepository runs read-only aggregate queries for the reports section.
type ReportRepository struct {
	pool *pgxpool.Pool
}

// NewReportRepository returns a ReportRepository bound to the given pool.
func NewReportRepository(pool *pgxpool.Pool) *ReportRepository {
	return &ReportRepository{pool: pool}
}

// TagSpend is a single row in a spend-by-tag report.
type TagSpend struct {
	TagName string
	Total   string // numeric string, e.g. "1234.56"
	Count   int
}

// MerchantMonthRow is one (merchant, month) bucket from the monthly merchant report.
type MerchantMonthRow struct {
	Identifier string
	Month      string // "YYYY-MM"
	Total      string
	MaxAmount  string
	AvgAmount  string
	Count      int
}

// SpendByPrimaryTag returns debit totals grouped by primary tag (merchants only).
func (r *ReportRepository) SpendByPrimaryTag(ctx context.Context) ([]TagSpend, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			tg.name,
			SUM(tx.amount)::text,
			COUNT(*)::int
		FROM transactions tx
		JOIN merchants m   ON m.identifier_name = tx.merchant_identifier
		JOIN tags      tg  ON tg.id = m.primary_tag_id
		WHERE tx.direction = 'debit'
		GROUP BY tg.name
		ORDER BY SUM(tx.amount) DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("spend by primary tag: %w", err)
	}
	defer rows.Close()

	var out []TagSpend
	for rows.Next() {
		var s TagSpend
		if err := rows.Scan(&s.TagName, &s.Total, &s.Count); err != nil {
			return nil, fmt.Errorf("scan primary tag row: %w", err)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// SpendBySecondaryTag returns debit totals grouped by secondary tag.
func (r *ReportRepository) SpendBySecondaryTag(ctx context.Context) ([]TagSpend, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			tg.name,
			SUM(tx.amount)::text,
			COUNT(*)::int
		FROM transactions tx
		JOIN merchants              m   ON m.identifier_name = tx.merchant_identifier
		JOIN merchant_secondary_tags mst ON mst.merchant_id = m.id
		JOIN tags                   tg  ON tg.id = mst.tag_id
		WHERE tx.direction = 'debit'
		GROUP BY tg.name
		ORDER BY SUM(tx.amount) DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("spend by secondary tag: %w", err)
	}
	defer rows.Close()

	var out []TagSpend
	for rows.Next() {
		var s TagSpend
		if err := rows.Scan(&s.TagName, &s.Total, &s.Count); err != nil {
			return nil, fmt.Errorf("scan secondary tag row: %w", err)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// DailySpend is a single day's debit total.
type DailySpend struct {
	Date  string // "YYYY-MM-DD"
	Total string
}

// DailySpend returns total debit spending per calendar day.
func (r *ReportRepository) DailySpend(ctx context.Context) ([]DailySpend, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT date::text, SUM(amount)::text
		FROM transactions
		WHERE direction = 'debit'
		GROUP BY date
		ORDER BY date
	`)
	if err != nil {
		return nil, fmt.Errorf("daily spend: %w", err)
	}
	defer rows.Close()

	var out []DailySpend
	for rows.Next() {
		var d DailySpend
		if err := rows.Scan(&d.Date, &d.Total); err != nil {
			return nil, fmt.Errorf("scan daily spend: %w", err)
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// MerchantsByMonth returns per-month aggregates for every known merchant identifier.
// Results are ordered by merchant identifier then month ascending.
func (r *ReportRepository) MerchantsByMonth(ctx context.Context) ([]MerchantMonthRow, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			tx.merchant_identifier,
			to_char(tx.date, 'YYYY-MM')       AS month,
			SUM(tx.amount)::text              AS total,
			MAX(tx.amount)::text              AS max_amount,
			ROUND(AVG(tx.amount), 2)::text    AS avg_amount,
			COUNT(*)::int                     AS tx_count
		FROM transactions tx
		WHERE tx.direction = 'debit'
		  AND tx.merchant_identifier IS NOT NULL
		GROUP BY tx.merchant_identifier, to_char(tx.date, 'YYYY-MM')
		ORDER BY tx.merchant_identifier, month
	`)
	if err != nil {
		return nil, fmt.Errorf("merchants by month: %w", err)
	}
	defer rows.Close()

	var out []MerchantMonthRow
	for rows.Next() {
		var row MerchantMonthRow
		if err := rows.Scan(
			&row.Identifier,
			&row.Month,
			&row.Total,
			&row.MaxAmount,
			&row.AvgAmount,
			&row.Count,
		); err != nil {
			return nil, fmt.Errorf("scan merchant month row: %w", err)
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

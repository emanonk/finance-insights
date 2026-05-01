package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/manoskammas/finance-insights/apps/api/internal/domain"
)

// ReportRepository runs read-only aggregate queries for the reports section.
type ReportRepository struct {
	pool *pgxpool.Pool
}

// NewReportRepository returns a ReportRepository bound to the given pool.
func NewReportRepository(pool *pgxpool.Pool) *ReportRepository {
	return &ReportRepository{pool: pool}
}

// accountFilter returns a WHERE fragment and args to optionally restrict to
// specific account IDs. argOffset is the number of existing $N args.
func accountFilter(accountIDs []string, argOffset int) (string, []any) {
	if len(accountIDs) == 0 {
		return "", nil
	}
	return fmt.Sprintf(" AND tx.account_id = ANY($%d)", argOffset+1), []any{accountIDs}
}

// SpendByPrimaryTag returns debit totals grouped by primary tag (merchants only).
func (r *ReportRepository) SpendByPrimaryTag(ctx context.Context, accountIDs []string) ([]domain.TagSpend, error) {
	filter, extra := accountFilter(accountIDs, 0)
	query := `
		SELECT
			tg.name,
			SUM(tx.amount)::text,
			COUNT(*)::int
		FROM transactions tx
		JOIN merchants m   ON m.identifier_name = tx.merchant_identifier
		JOIN tags      tg  ON tg.id = m.primary_tag_id
		WHERE tx.direction = 'debit'` + filter + `
		GROUP BY tg.name
		ORDER BY SUM(tx.amount) DESC`

	rows, err := r.pool.Query(ctx, query, extra...)
	if err != nil {
		return nil, fmt.Errorf("spend by primary tag: %w", err)
	}
	defer rows.Close()

	var out []domain.TagSpend
	for rows.Next() {
		var s domain.TagSpend
		if err := rows.Scan(&s.TagName, &s.Total, &s.Count); err != nil {
			return nil, fmt.Errorf("scan primary tag row: %w", err)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// SpendBySecondaryTag returns debit totals grouped by secondary tag.
func (r *ReportRepository) SpendBySecondaryTag(ctx context.Context, accountIDs []string) ([]domain.TagSpend, error) {
	filter, extra := accountFilter(accountIDs, 0)
	query := `
		SELECT
			tg.name,
			SUM(tx.amount)::text,
			COUNT(*)::int
		FROM transactions tx
		JOIN merchants              m   ON m.identifier_name = tx.merchant_identifier
		JOIN merchant_secondary_tags mst ON mst.merchant_id = m.id
		JOIN tags                   tg  ON tg.id = mst.tag_id
		WHERE tx.direction = 'debit'` + filter + `
		GROUP BY tg.name
		ORDER BY SUM(tx.amount) DESC`

	rows, err := r.pool.Query(ctx, query, extra...)
	if err != nil {
		return nil, fmt.Errorf("spend by secondary tag: %w", err)
	}
	defer rows.Close()

	var out []domain.TagSpend
	for rows.Next() {
		var s domain.TagSpend
		if err := rows.Scan(&s.TagName, &s.Total, &s.Count); err != nil {
			return nil, fmt.Errorf("scan secondary tag row: %w", err)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// DailySpend returns total debit spending per calendar day.
func (r *ReportRepository) DailySpend(ctx context.Context, accountIDs []string) ([]domain.DailySpend, error) {
	filter, extra := accountFilter(accountIDs, 0)
	// DailySpend doesn't join merchants, so we need a bare tx alias.
	bareFilter := strings.ReplaceAll(filter, "tx.account_id", "account_id")
	query := `
		SELECT date::text, SUM(amount)::text
		FROM transactions
		WHERE direction = 'debit'` + bareFilter + `
		GROUP BY date
		ORDER BY date`

	rows, err := r.pool.Query(ctx, query, extra...)
	if err != nil {
		return nil, fmt.Errorf("daily spend: %w", err)
	}
	defer rows.Close()

	var out []domain.DailySpend
	for rows.Next() {
		var d domain.DailySpend
		if err := rows.Scan(&d.Date, &d.Total); err != nil {
			return nil, fmt.Errorf("scan daily spend: %w", err)
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// MerchantsByMonth returns per-month aggregates for every known merchant identifier.
// Results are ordered by merchant identifier then month ascending.
func (r *ReportRepository) MerchantsByMonth(ctx context.Context, accountIDs []string) ([]domain.MerchantMonthRow, error) {
	filter, extra := accountFilter(accountIDs, 0)
	query := `
		SELECT
			tx.merchant_identifier,
			to_char(tx.date, 'YYYY-MM')       AS month,
			SUM(tx.amount)::text              AS total,
			MAX(tx.amount)::text              AS max_amount,
			ROUND(AVG(tx.amount), 2)::text    AS avg_amount,
			COUNT(*)::int                     AS tx_count
		FROM transactions tx
		WHERE tx.direction = 'debit'
		  AND tx.merchant_identifier IS NOT NULL` + filter + `
		GROUP BY tx.merchant_identifier, to_char(tx.date, 'YYYY-MM')
		ORDER BY tx.merchant_identifier, month`

	rows, err := r.pool.Query(ctx, query, extra...)
	if err != nil {
		return nil, fmt.Errorf("merchants by month: %w", err)
	}
	defer rows.Close()

	var out []domain.MerchantMonthRow
	for rows.Next() {
		var row domain.MerchantMonthRow
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

// RecurringByYear returns recurring debit patterns with at least 10 occurrences
// in a full calendar year (same merchant + same rounded amount).
func (r *ReportRepository) RecurringByYear(ctx context.Context) ([]domain.RecurringCharge, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			EXTRACT(YEAR FROM t.date)::int AS calendar_year,
			t.merchant_identifier,
			round(t.amount::numeric, 2)::text AS charge_amount,
			COUNT(*)::int AS occurrences,
			SUM(t.amount)::text AS total_debited,
			MIN(t.date)::text AS first_date,
			MAX(t.date)::text AS last_date
		FROM transactions t
		LEFT JOIN transaction_extra te ON te.transaction_id = t.id
		WHERE t.direction = 'debit'
		  AND t.merchant_identifier IS NOT NULL
		  AND COALESCE(te.in_report, true)
		GROUP BY calendar_year, t.merchant_identifier, round(t.amount::numeric, 2)
		HAVING COUNT(*) >= 10
		ORDER BY calendar_year, SUM(t.amount) DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("recurring by year: %w", err)
	}
	defer rows.Close()
	return scanRecurringRows(rows)
}

// RecurringYTD returns recurring debit patterns with at least 3 occurrences,
// suitable for detecting recurring charges in a partial year.
func (r *ReportRepository) RecurringYTD(ctx context.Context) ([]domain.RecurringCharge, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			EXTRACT(YEAR FROM t.date)::int AS calendar_year,
			t.merchant_identifier,
			round(t.amount::numeric, 2)::text AS charge_amount,
			COUNT(*)::int AS occurrences,
			SUM(t.amount)::text AS total_debited,
			MIN(t.date)::text AS first_date,
			MAX(t.date)::text AS last_date
		FROM transactions t
		LEFT JOIN transaction_extra te ON te.transaction_id = t.id
		WHERE t.direction = 'debit'
		  AND t.merchant_identifier IS NOT NULL
		  AND COALESCE(te.in_report, true)
		GROUP BY calendar_year, t.merchant_identifier, round(t.amount::numeric, 2)
		HAVING COUNT(*) >= 3
		ORDER BY calendar_year, SUM(t.amount) DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("recurring ytd: %w", err)
	}
	defer rows.Close()
	return scanRecurringRows(rows)
}

func scanRecurringRows(rows interface {
	Next() bool
	Scan(...any) error
	Err() error
}) ([]domain.RecurringCharge, error) {
	var out []domain.RecurringCharge
	for rows.Next() {
		var c domain.RecurringCharge
		if err := rows.Scan(
			&c.CalendarYear,
			&c.Identifier,
			&c.Amount,
			&c.Occurrences,
			&c.TotalDebited,
			&c.FirstDate,
			&c.LastDate,
		); err != nil {
			return nil, fmt.Errorf("scan recurring charge: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

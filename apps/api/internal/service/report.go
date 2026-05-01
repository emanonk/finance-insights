package service

import (
	"context"
	"fmt"

	"github.com/manoskammas/finance-insights/apps/api/internal/domain"
)

type reportStore interface {
	SpendByPrimaryTag(ctx context.Context, accountIDs []string) ([]domain.TagSpend, error)
	SpendBySecondaryTag(ctx context.Context, accountIDs []string) ([]domain.TagSpend, error)
	MerchantsByMonth(ctx context.Context, accountIDs []string) ([]domain.MerchantMonthRow, error)
	DailySpend(ctx context.Context, accountIDs []string) ([]domain.DailySpend, error)
}

// Report serves aggregate report queries.
type Report struct {
	repo reportStore
}

// NewReport constructs a Report service.
func NewReport(repo reportStore) *Report {
	return &Report{repo: repo}
}

// SpendByPrimaryTag returns debit totals by primary tag, descending by total.
func (s *Report) SpendByPrimaryTag(ctx context.Context, accountIDs []string) ([]domain.TagSpend, error) {
	rows, err := s.repo.SpendByPrimaryTag(ctx, accountIDs)
	if err != nil {
		return nil, fmt.Errorf("spend by primary tag: %w", err)
	}
	return rows, nil
}

// SpendBySecondaryTag returns debit totals by secondary tag, descending by total.
func (s *Report) SpendBySecondaryTag(ctx context.Context, accountIDs []string) ([]domain.TagSpend, error) {
	rows, err := s.repo.SpendBySecondaryTag(ctx, accountIDs)
	if err != nil {
		return nil, fmt.Errorf("spend by secondary tag: %w", err)
	}
	return rows, nil
}

// MerchantMonthly is the monthly breakdown for a single merchant.
type MerchantMonthly struct {
	Month     string
	Total     string
	MaxAmount string
	AvgAmount string
	Count     int
}

// MerchantSummary aggregates all monthly rows for a merchant, sorted month-asc.
type MerchantSummary struct {
	Identifier string
	TotalSpend string
	TxCount    int
	Months     []MerchantMonthly
}

// DailySpend returns total debit spending per calendar day.
func (s *Report) DailySpend(ctx context.Context, accountIDs []string) ([]domain.DailySpend, error) {
	rows, err := s.repo.DailySpend(ctx, accountIDs)
	if err != nil {
		return nil, fmt.Errorf("daily spend: %w", err)
	}
	return rows, nil
}

// MerchantsByMonth returns per-merchant monthly aggregates, ordered by total
// spend descending so the highest-spending merchants come first.
func (s *Report) MerchantsByMonth(ctx context.Context, accountIDs []string) ([]MerchantSummary, error) {
	flat, err := s.repo.MerchantsByMonth(ctx, accountIDs)
	if err != nil {
		return nil, fmt.Errorf("merchants by month: %w", err)
	}

	indexMap := map[string]int{}
	var summaries []MerchantSummary

	for _, row := range flat {
		idx, ok := indexMap[row.Identifier]
		if !ok {
			idx = len(summaries)
			indexMap[row.Identifier] = idx
			summaries = append(summaries, MerchantSummary{
				Identifier: row.Identifier,
			})
		}
		s := &summaries[idx]
		s.Months = append(s.Months, MerchantMonthly{
			Month:     row.Month,
			Total:     row.Total,
			MaxAmount: row.MaxAmount,
			AvgAmount: row.AvgAmount,
			Count:     row.Count,
		})
		s.TxCount += row.Count
	}

	for i := range summaries {
		summaries[i].TotalSpend = sumStrings(summaries[i].Months)
	}
	sortMerchantsByTotal(summaries)

	return summaries, nil
}

func sumStrings(months []MerchantMonthly) string {
	var total float64
	for _, m := range months {
		var v float64
		fmt.Sscanf(m.Total, "%f", &v) //nolint:errcheck
		total += v
	}
	return fmt.Sprintf("%.2f", total)
}

func sortMerchantsByTotal(ss []MerchantSummary) {
	for i := 1; i < len(ss); i++ {
		for j := i; j > 0; j-- {
			var a, b float64
			fmt.Sscanf(ss[j-1].TotalSpend, "%f", &a) //nolint:errcheck
			fmt.Sscanf(ss[j].TotalSpend, "%f", &b)   //nolint:errcheck
			if b > a {
				ss[j-1], ss[j] = ss[j], ss[j-1]
			} else {
				break
			}
		}
	}
}

package handler

import (
	"context"
	"net/http"

	"github.com/manoskammas/finance-insights/apps/api/internal/domain"
	"github.com/manoskammas/finance-insights/apps/api/internal/service"
)

type reportService interface {
	SpendByPrimaryTag(ctx context.Context) ([]domain.TagSpend, error)
	SpendBySecondaryTag(ctx context.Context) ([]domain.TagSpend, error)
	MerchantsByMonth(ctx context.Context) ([]service.MerchantSummary, error)
	DailySpend(ctx context.Context) ([]domain.DailySpend, error)
}

// Report serves report endpoints.
type Report struct {
	Service reportService
}

// ── DTOs ──────────────────────────────────────────────────────────────────────

type tagSpendDTO struct {
	TagName string `json:"tagName"`
	Total   string `json:"total"`
	Count   int    `json:"count"`
}

type tagSpendResponse struct {
	Items []tagSpendDTO `json:"items"`
}

type merchantMonthlyDTO struct {
	Month     string `json:"month"`
	Total     string `json:"total"`
	MaxAmount string `json:"maxAmount"`
	AvgAmount string `json:"avgAmount"`
	Count     int    `json:"count"`
}

type merchantSummaryDTO struct {
	Identifier string               `json:"identifier"`
	TotalSpend string               `json:"totalSpend"`
	TxCount    int                  `json:"txCount"`
	Months     []merchantMonthlyDTO `json:"months"`
}

type merchantsByMonthResponse struct {
	Items []merchantSummaryDTO `json:"items"`
}

// ── Handlers ──────────────────────────────────────────────────────────────────

// SpendByPrimaryTag handles GET /reports/spend-by-primary-tag.
func (h *Report) SpendByPrimaryTag(w http.ResponseWriter, r *http.Request) {
	rows, err := h.Service.SpendByPrimaryTag(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load primary tag report")
		return
	}
	items := make([]tagSpendDTO, 0, len(rows))
	for _, row := range rows {
		items = append(items, tagSpendDTO{TagName: row.TagName, Total: row.Total, Count: row.Count})
	}
	writeJSON(w, http.StatusOK, tagSpendResponse{Items: items})
}

// SpendBySecondaryTag handles GET /reports/spend-by-secondary-tag.
func (h *Report) SpendBySecondaryTag(w http.ResponseWriter, r *http.Request) {
	rows, err := h.Service.SpendBySecondaryTag(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load secondary tag report")
		return
	}
	items := make([]tagSpendDTO, 0, len(rows))
	for _, row := range rows {
		items = append(items, tagSpendDTO{TagName: row.TagName, Total: row.Total, Count: row.Count})
	}
	writeJSON(w, http.StatusOK, tagSpendResponse{Items: items})
}

type dailySpendDTO struct {
	Date  string `json:"date"`
	Total string `json:"total"`
}

type dailySpendResponse struct {
	Items []dailySpendDTO `json:"items"`
}

// DailySpend handles GET /reports/daily-spend.
func (h *Report) DailySpend(w http.ResponseWriter, r *http.Request) {
	rows, err := h.Service.DailySpend(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load daily spend")
		return
	}
	items := make([]dailySpendDTO, 0, len(rows))
	for _, row := range rows {
		items = append(items, dailySpendDTO{Date: row.Date, Total: row.Total})
	}
	writeJSON(w, http.StatusOK, dailySpendResponse{Items: items})
}

// MerchantsByMonth handles GET /reports/merchants-by-month.
func (h *Report) MerchantsByMonth(w http.ResponseWriter, r *http.Request) {
	summaries, err := h.Service.MerchantsByMonth(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load merchant monthly report")
		return
	}
	items := make([]merchantSummaryDTO, 0, len(summaries))
	for _, ms := range summaries {
		months := make([]merchantMonthlyDTO, 0, len(ms.Months))
		for _, m := range ms.Months {
			months = append(months, merchantMonthlyDTO{
				Month:     m.Month,
				Total:     m.Total,
				MaxAmount: m.MaxAmount,
				AvgAmount: m.AvgAmount,
				Count:     m.Count,
			})
		}
		items = append(items, merchantSummaryDTO{
			Identifier: ms.Identifier,
			TotalSpend: ms.TotalSpend,
			TxCount:    ms.TxCount,
			Months:     months,
		})
	}
	writeJSON(w, http.StatusOK, merchantsByMonthResponse{Items: items})
}

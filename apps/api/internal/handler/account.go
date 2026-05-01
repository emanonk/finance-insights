package handler

import (
	"context"
	"net/http"

	"github.com/manoskammas/finance-insights/apps/api/internal/domain"
)

type accountService interface {
	List(ctx context.Context) ([]domain.Account, error)
}

// Account serves account endpoints.
type Account struct {
	Service accountService
}

type accountDTO struct {
	ID            string `json:"id"`
	BankName      string `json:"bankName"`
	AccountNumber string `json:"accountNumber"`
}

type accountListResponse struct {
	Items []accountDTO `json:"items"`
}

// List handles GET /accounts.
func (h *Account) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.Service.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list accounts")
		return
	}
	dtos := make([]accountDTO, 0, len(items))
	for _, a := range items {
		dtos = append(dtos, accountDTO{ID: a.ID, BankName: a.BankName, AccountNumber: a.AccountNumber})
	}
	writeJSON(w, http.StatusOK, accountListResponse{Items: dtos})
}

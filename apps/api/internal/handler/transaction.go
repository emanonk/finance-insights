package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/manoskammas/finance-insights/apps/api/internal/repository"
	"github.com/manoskammas/finance-insights/apps/api/internal/service"
)

// transactionLister is the subset of service.Transaction used by the handler.
type transactionLister interface {
	List(ctx context.Context, limit, offset int) (service.ListResult, error)
}

// Transaction serves transaction read endpoints.
type Transaction struct {
	Service transactionLister
}

// transactionDTO is the transport shape of a transaction.
type transactionDTO struct {
	ID                      string  `json:"id"`
	AccountID               int64   `json:"accountId"`
	Date                    string  `json:"date"`
	BankReferenceNumber     *string `json:"bankReferenceNumber"`
	Justification           *string `json:"justification"`
	Indicator               *string `json:"indicator"`
	MerchantIdentifier      *string `json:"merchantIdentifier"`
	Amount1                 *string `json:"amount1"`
	MCCCode                 *string `json:"mccCode"`
	CardMasked              *string `json:"cardMasked"`
	Reference               *string `json:"reference"`
	Description             string  `json:"description"`
	PaymentMethod           *string `json:"paymentMethod"`
	Direction               string  `json:"direction"`
	Amount                  string  `json:"amount"`
	BalanceAfterTransaction *string `json:"balanceAfterTransaction"`
	StatementFileName       *string `json:"statementFileName"`
}

// transactionListResponse is the response envelope for GET /transactions.
type transactionListResponse struct {
	Items  []transactionDTO `json:"items"`
	Total  int              `json:"total"`
	Limit  int              `json:"limit"`
	Offset int              `json:"offset"`
}

// List handles GET /transactions.
func (h *Transaction) List(w http.ResponseWriter, r *http.Request) {
	limit, err := parseIntQuery(r, "limit", 0)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid limit")
		return
	}
	offset, err := parseIntQuery(r, "offset", 0)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid offset")
		return
	}

	result, err := h.Service.List(r.Context(), limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list transactions")
		return
	}

	items := make([]transactionDTO, 0, len(result.Items))
	for _, t := range result.Items {
		items = append(items, toTransactionDTO(t))
	}
	writeJSON(w, http.StatusOK, transactionListResponse{
		Items:  items,
		Total:  result.Total,
		Limit:  result.Limit,
		Offset: result.Offset,
	})
}

func parseIntQuery(r *http.Request, key string, defaultVal int) (int, error) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return defaultVal, nil
	}
	return strconv.Atoi(raw)
}

func toTransactionDTO(t repository.Transaction) transactionDTO {
	desc := ""
	if t.Description != nil {
		desc = *t.Description
	}
	return transactionDTO{
		ID:                      fmt.Sprintf("%d", t.ID),
		AccountID:               t.AccountID,
		Date:                    t.Date.Format("2006-01-02"),
		BankReferenceNumber:     t.BankReferenceNumber,
		Justification:           t.Justification,
		Indicator:               t.Indicator,
		MerchantIdentifier:      t.MerchantIdentifier,
		Amount1:                 t.Amount1,
		MCCCode:                 t.MCCCode,
		CardMasked:              t.CardMasked,
		Reference:               t.Reference,
		Description:             desc,
		PaymentMethod:           t.PaymentMethod,
		Direction:               t.Direction,
		Amount:                  t.Amount,
		BalanceAfterTransaction: t.BalanceAfterTransaction,
		StatementFileName:       t.StatementFileName,
	}
}

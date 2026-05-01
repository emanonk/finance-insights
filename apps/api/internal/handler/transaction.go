package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/manoskammas/finance-insights/apps/api/internal/domain"
	"github.com/manoskammas/finance-insights/apps/api/internal/service"
)

// transactionLister is the subset of service.Transaction used by the handler.
type transactionLister interface {
	List(ctx context.Context, limit, offset int, accountIDs []string) (service.ListResult, error)
}

// Transaction serves transaction read endpoints.
type Transaction struct {
	Service transactionLister
}

// transactionDTO is the transport shape of a transaction.
type transactionDTO struct {
	ID                   string   `json:"id"`
	AccountID            string   `json:"accountId"`
	Date                 string   `json:"date"`
	BankReference        *string  `json:"bankReference"`
	TransactionReference *string  `json:"transactionReference"`
	MerchantIdentifier   *string  `json:"merchantIdentifier"`
	BalanceBefore        int      `json:"balanceBefore"`
	BalanceAfter         int      `json:"balanceAfter"`
	Amount               int      `json:"amount"`
	Direction            string   `json:"direction"`
	RawData              []string `json:"rawData"`
	StatementFileName    *string  `json:"statementFileName"`
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

	var accountIDs []string
	if raw := r.URL.Query().Get("accountIds"); raw != "" {
		for _, id := range strings.Split(raw, ",") {
			if id = strings.TrimSpace(id); id != "" {
				accountIDs = append(accountIDs, id)
			}
		}
	}

	result, err := h.Service.List(r.Context(), limit, offset, accountIDs)
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

func toTransactionDTO(t domain.Transaction) transactionDTO {
	return transactionDTO{
		ID:                   fmt.Sprintf("%d", t.ID),
		AccountID:            t.AccountID,
		Date:                 t.Date.Format("2006-01-02"),
		BankReference:        t.BankReference,
		TransactionReference: t.TransactionReference,
		MerchantIdentifier:   t.MerchantIdentifier,
		BalanceBefore:        t.BalanceBefore,
		BalanceAfter:         t.BalanceAfter,
		Amount:               t.Amount,
		Direction:            t.Direction,
		RawData:              t.RawData,
		StatementFileName:    t.StatementFileName,
	}
}

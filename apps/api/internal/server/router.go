package server

import (
	"net/http"

	"github.com/manoskammas/finance-insights/apps/api/internal/handler"
	"github.com/manoskammas/finance-insights/apps/api/internal/service"
)

// Deps bundles the dependencies the router wires into handlers.
type Deps struct {
	System      *service.System
	Statement   *service.Statement
	Account     *service.Account
	Transaction *service.Transaction
	Merchant    *service.Merchant
	Report      *service.Report
}

// NewRouter builds the HTTP handler graph for the API.
func NewRouter(d Deps) http.Handler {
	mux := http.NewServeMux()

	health := &handler.Health{System: d.System}
	mux.HandleFunc("GET /health", health.Get)

	statement := &handler.Statement{Service: d.Statement}
	mux.HandleFunc("POST /statements", statement.Create)

	account := &handler.Account{Service: d.Account}
	mux.HandleFunc("GET /accounts", account.List)

	transaction := &handler.Transaction{Service: d.Transaction}
	mux.HandleFunc("GET /transactions", transaction.List)

	merchant := &handler.Merchant{Service: d.Merchant}
	mux.HandleFunc("GET /merchants/top", merchant.TopIdentifiers)
	mux.HandleFunc("POST /merchants", merchant.Upsert)

	report := &handler.Report{Service: d.Report}
	mux.HandleFunc("GET /reports/spend-by-primary-tag", report.SpendByPrimaryTag)
	mux.HandleFunc("GET /reports/spend-by-secondary-tag", report.SpendBySecondaryTag)
	mux.HandleFunc("GET /reports/merchants-by-month", report.MerchantsByMonth)
	mux.HandleFunc("GET /reports/daily-spend", report.DailySpend)

	return mux
}

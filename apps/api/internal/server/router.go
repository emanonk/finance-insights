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
	Transaction *service.Transaction
}

// NewRouter builds the HTTP handler graph for the API.
func NewRouter(d Deps) http.Handler {
	mux := http.NewServeMux()

	health := &handler.Health{System: d.System}
	mux.HandleFunc("GET /health", health.Get)

	statement := &handler.Statement{Service: d.Statement}
	mux.HandleFunc("POST /statements", statement.Create)

	transaction := &handler.Transaction{Service: d.Transaction}
	mux.HandleFunc("GET /transactions", transaction.List)

	return mux
}

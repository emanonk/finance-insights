package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/manoskammas/finance-insights/apps/api/internal/repository"
)

type merchantService interface {
	TopIdentifiers(ctx context.Context, limit int) ([]repository.IdentifierCount, error)
	UpsertMerchant(ctx context.Context, identifierName, primaryTagName string, secondaryTagNames []string) (*repository.Merchant, error)
}

// Merchant serves merchant and tag endpoints.
type Merchant struct {
	Service merchantService
}

type tagDTO struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type merchantDTO struct {
	ID             int64    `json:"id"`
	IdentifierName string   `json:"identifierName"`
	PrimaryTag     tagDTO   `json:"primaryTag"`
	SecondaryTags  []tagDTO `json:"secondaryTags"`
}

type identifierCountDTO struct {
	Identifier string       `json:"identifier"`
	Count      int          `json:"count"`
	Merchant   *merchantDTO `json:"merchant"`
}

type topIdentifiersResponse struct {
	Items []identifierCountDTO `json:"items"`
}

// TopIdentifiers handles GET /merchants/top.
func (h *Merchant) TopIdentifiers(w http.ResponseWriter, r *http.Request) {
	limit := 30
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil {
			limit = v
		}
	}

	items, err := h.Service.TopIdentifiers(r.Context(), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load top identifiers")
		return
	}

	dtos := make([]identifierCountDTO, 0, len(items))
	for _, ic := range items {
		dto := identifierCountDTO{Identifier: ic.Identifier, Count: ic.Count}
		if ic.Merchant != nil {
			dto.Merchant = toMerchantDTO(ic.Merchant)
		}
		dtos = append(dtos, dto)
	}
	writeJSON(w, http.StatusOK, topIdentifiersResponse{Items: dtos})
}

type upsertMerchantRequest struct {
	IdentifierName    string   `json:"identifierName"`
	PrimaryTagName    string   `json:"primaryTagName"`
	SecondaryTagNames []string `json:"secondaryTagNames"`
}

// Upsert handles POST /merchants.
func (h *Merchant) Upsert(w http.ResponseWriter, r *http.Request) {
	var req upsertMerchantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.IdentifierName = strings.TrimSpace(req.IdentifierName)
	req.PrimaryTagName = strings.TrimSpace(req.PrimaryTagName)

	if req.IdentifierName == "" {
		writeError(w, http.StatusBadRequest, "identifierName is required")
		return
	}
	if req.PrimaryTagName == "" {
		writeError(w, http.StatusBadRequest, "primaryTagName is required")
		return
	}

	// Trim whitespace from secondary tag names and drop empty strings.
	cleaned := req.SecondaryTagNames[:0]
	for _, s := range req.SecondaryTagNames {
		if t := strings.TrimSpace(s); t != "" {
			cleaned = append(cleaned, t)
		}
	}

	m, err := h.Service.UpsertMerchant(r.Context(), req.IdentifierName, req.PrimaryTagName, cleaned)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save merchant")
		return
	}
	writeJSON(w, http.StatusOK, toMerchantDTO(m))
}

func toMerchantDTO(m *repository.Merchant) *merchantDTO {
	sec := make([]tagDTO, 0, len(m.SecondaryTags))
	for _, t := range m.SecondaryTags {
		sec = append(sec, tagDTO{ID: t.ID, Name: t.Name, Type: t.Type})
	}
	return &merchantDTO{
		ID:             m.ID,
		IdentifierName: m.IdentifierName,
		PrimaryTag:     tagDTO{ID: m.PrimaryTag.ID, Name: m.PrimaryTag.Name, Type: m.PrimaryTag.Type},
		SecondaryTags:  sec,
	}
}

package handler

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/manoskammas/finance-insights/apps/api/internal/service"
)

const (
	maxStatementUploadBytes = 25 << 20 // 25 MiB
	statementFormField      = "file"
	bankNameFormField       = "bank_name"
)

// statementIngester is the subset of service.Statement used by the handler.
type statementIngester interface {
	Ingest(ctx context.Context, bankName string, fileName string, r io.Reader) (service.IngestResult, error)
}

// Statement serves the statement upload endpoint.
type Statement struct {
	Service statementIngester
}

// statementCreateResponse is the transport payload returned on successful upload.
type statementCreateResponse struct {
	FileName         string `json:"fileName"`
	TransactionCount int    `json:"transactionCount"`
}

// Create handles POST /statements.
//
// Required form fields:
//   - file      — the PDF bank statement (multipart)
//   - bank_name — identifier of the bank (e.g. "piraeus")
func (h *Statement) Create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxStatementUploadBytes)

	if err := r.ParseMultipartForm(maxStatementUploadBytes); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeError(w, http.StatusRequestEntityTooLarge, "file exceeds 25 MiB limit")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}

	bankName := strings.ToLower(strings.TrimSpace(r.FormValue(bankNameFormField)))
	if bankName == "" {
		writeError(w, http.StatusBadRequest, "missing bank_name field")
		return
	}

	file, header, err := r.FormFile(statementFormField)
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing file field")
		return
	}
	defer file.Close()

	if !isPDF(header.Filename, header.Header.Get("Content-Type")) {
		writeError(w, http.StatusUnsupportedMediaType, "only PDF files are accepted")
		return
	}

	result, err := h.Service.Ingest(r.Context(), bankName, header.Filename, file)
	if err != nil {
		slog.Error("statement ingest", "err", err, "file", header.Filename)
		writeError(w, http.StatusInternalServerError, "failed to ingest statement")
		return
	}

	writeJSON(w, http.StatusCreated, statementCreateResponse{
		FileName:         result.FileName,
		TransactionCount: result.TransactionCount,
	})
}

func isPDF(fileName, contentType string) bool {
	if strings.ToLower(filepath.Ext(fileName)) != ".pdf" {
		return false
	}
	if contentType == "" {
		return true
	}
	// strip any parameters like "; charset=..."
	if i := strings.IndexByte(contentType, ';'); i >= 0 {
		contentType = contentType[:i]
	}
	ct := strings.ToLower(strings.TrimSpace(contentType))
	return ct == "application/pdf" || ct == "application/octet-stream"
}

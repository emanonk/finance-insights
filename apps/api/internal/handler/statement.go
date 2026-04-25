package handler

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/manoskammas/finance-insights/apps/api/internal/service"
)

const (
	maxStatementUploadBytes = 25 << 20 // 25 MiB
	statementFormField      = "file"
	accountIDFormField      = "account_id"
)

// statementIngester is the subset of service.Statement used by the handler.
type statementIngester interface {
	Ingest(ctx context.Context, accountID int64, fileName string, r io.Reader) (service.IngestResult, error)
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
//   - file       — the PDF bank statement (multipart)
//   - account_id — integer FK referencing the target account
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

	rawAccountID := r.FormValue(accountIDFormField)
	if rawAccountID == "" {
		writeError(w, http.StatusBadRequest, "missing account_id field")
		return
	}
	accountID, err := strconv.ParseInt(rawAccountID, 10, 64)
	if err != nil || accountID <= 0 {
		writeError(w, http.StatusBadRequest, "account_id must be a positive integer")
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

	result, err := h.Service.Ingest(r.Context(), accountID, header.Filename, file)
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

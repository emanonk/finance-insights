package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"strings"
	"testing"

	"github.com/manoskammas/finance-insights/apps/api/internal/service"
)

type fakeStatementService struct {
	gotAccountID int64
	gotFileName  string
	gotBody      []byte
	result       service.IngestResult
	err          error
}

func (f *fakeStatementService) Ingest(_ context.Context, accountID int64, fileName string, r io.Reader) (service.IngestResult, error) {
	f.gotAccountID = accountID
	f.gotFileName = fileName
	body, err := io.ReadAll(r)
	if err != nil {
		return service.IngestResult{}, err
	}
	f.gotBody = body
	if f.err != nil {
		return service.IngestResult{}, f.err
	}
	return f.result, nil
}

func buildMultipart(t *testing.T, fieldName, fileName, contentType string, body []byte) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition",
		`form-data; name="`+fieldName+`"; filename="`+fileName+`"`)
	if contentType != "" {
		header.Set("Content-Type", contentType)
	}
	part, err := w.CreatePart(header)
	if err != nil {
		t.Fatalf("create part: %v", err)
	}
	if _, err := part.Write(body); err != nil {
		t.Fatalf("write part: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	return &buf, w.FormDataContentType()
}

// buildMultipartWithAccountID builds a multipart form that includes the file
// part and an account_id text field.
func buildMultipartWithAccountID(t *testing.T, accountID, fileName, contentType string, body []byte) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	// account_id text field
	if err := w.WriteField(accountIDFormField, accountID); err != nil {
		t.Fatalf("write account_id field: %v", err)
	}

	// file part
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition",
		`form-data; name="file"; filename="`+fileName+`"`)
	if contentType != "" {
		header.Set("Content-Type", contentType)
	}
	part, err := w.CreatePart(header)
	if err != nil {
		t.Fatalf("create part: %v", err)
	}
	if _, err := part.Write(body); err != nil {
		t.Fatalf("write part: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	return &buf, w.FormDataContentType()
}

func TestStatementHandler_Create_HappyPath(t *testing.T) {
	t.Parallel()

	svc := &fakeStatementService{
		result: service.IngestResult{
			FileName:         "statement.pdf",
			TransactionCount: 3,
		},
	}
	h := &Statement{Service: svc}

	body, contentType := buildMultipartWithAccountID(t, "1", "statement.pdf", "application/pdf", []byte("%PDF-1.4\n%%EOF"))
	req := httptest.NewRequest(http.MethodPost, "/statements", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}

	var resp statementCreateResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.FileName != "statement.pdf" {
		t.Errorf("fileName = %q", resp.FileName)
	}
	if resp.TransactionCount != 3 {
		t.Errorf("txCount = %d, want 3", resp.TransactionCount)
	}
	if svc.gotFileName != "statement.pdf" {
		t.Errorf("service fileName = %q", svc.gotFileName)
	}
	if svc.gotAccountID != 1 {
		t.Errorf("service accountID = %d, want 1", svc.gotAccountID)
	}
}

func TestStatementHandler_Create_MissingAccountID(t *testing.T) {
	t.Parallel()

	svc := &fakeStatementService{}
	h := &Statement{Service: svc}

	// build form without account_id
	body, contentType := buildMultipart(t, "file", "statement.pdf", "application/pdf", []byte("%PDF"))
	req := httptest.NewRequest(http.MethodPost, "/statements", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestStatementHandler_Create_InvalidAccountID(t *testing.T) {
	t.Parallel()

	svc := &fakeStatementService{}
	h := &Statement{Service: svc}

	body, contentType := buildMultipartWithAccountID(t, "abc", "statement.pdf", "application/pdf", []byte("%PDF"))
	req := httptest.NewRequest(http.MethodPost, "/statements", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestStatementHandler_Create_MissingFileField(t *testing.T) {
	t.Parallel()

	svc := &fakeStatementService{}
	h := &Statement{Service: svc}

	body, contentType := buildMultipart(t, "other", "x.pdf", "application/pdf", []byte("x"))
	req := httptest.NewRequest(http.MethodPost, "/statements", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestStatementHandler_Create_NonPDFRejected(t *testing.T) {
	t.Parallel()

	svc := &fakeStatementService{}
	h := &Statement{Service: svc}

	body, contentType := buildMultipartWithAccountID(t, "1", "notes.txt", "text/plain", []byte("hello"))
	req := httptest.NewRequest(http.MethodPost, "/statements", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("status = %d, want 415", rec.Code)
	}
}

func TestStatementHandler_Create_NonMultipartRejected(t *testing.T) {
	t.Parallel()

	svc := &fakeStatementService{}
	h := &Statement{Service: svc}

	req := httptest.NewRequest(http.MethodPost, "/statements", strings.NewReader("raw body"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestStatementHandler_Create_ServiceErrorReturns500(t *testing.T) {
	t.Parallel()

	svc := &fakeStatementService{err: errors.New("parser down")}
	h := &Statement{Service: svc}

	body, contentType := buildMultipartWithAccountID(t, "1", "s.pdf", "application/pdf", []byte("%PDF"))
	req := httptest.NewRequest(http.MethodPost, "/statements", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
}

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
	gotBankName string
	gotFileName string
	gotBody     []byte
	result      service.IngestResult
	err         error
}

func (f *fakeStatementService) Ingest(_ context.Context, bankName string, fileName string, r io.Reader) (service.IngestResult, error) {
	f.gotBankName = bankName
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

// buildMultipartWithBankName builds a multipart form that includes the file
// part and a bank_name text field.
func buildMultipartWithBankName(t *testing.T, bankName, fileName, contentType string, body []byte) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	if err := w.WriteField(bankNameFormField, bankName); err != nil {
		t.Fatalf("write bank_name field: %v", err)
	}

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

	body, contentType := buildMultipartWithBankName(t, "piraeus", "statement.pdf", "application/pdf", []byte("%PDF-1.4\n%%EOF"))
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
	if svc.gotBankName != "piraeus" {
		t.Errorf("service bankName = %q, want piraeus", svc.gotBankName)
	}
}

func TestStatementHandler_Create_MissingBankName(t *testing.T) {
	t.Parallel()

	svc := &fakeStatementService{}
	h := &Statement{Service: svc}

	// build form without bank_name
	body, contentType := buildMultipart(t, "file", "statement.pdf", "application/pdf", []byte("%PDF"))
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

	body, contentType := buildMultipartWithBankName(t, "piraeus", "notes.txt", "text/plain", []byte("hello"))
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

	body, contentType := buildMultipartWithBankName(t, "piraeus", "s.pdf", "application/pdf", []byte("%PDF"))
	req := httptest.NewRequest(http.MethodPost, "/statements", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
}

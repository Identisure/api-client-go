package idkollen

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"
)

type DocumentUploadResponse struct {
	Id   string `json:"id"`
	Hash string `json:"hash"`
}

type DocumentEndpoint struct {
	client *IdkollenClient
}

// Escapes the characters that would break the quoted Content-Disposition value,
// using the WHATWG multipart/form-data encoding browsers use.
var formFieldEscaper = strings.NewReplacer(`"`, "%22", "\r", "%0D", "\n", "%0A")

func (e *DocumentEndpoint) Upload(ctx context.Context, data []byte, filename string, mimeType string) (*DocumentUploadResponse, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="file"; filename="`+formFieldEscaper.Replace(filename)+`"`)
	h.Set("Content-Type", mimeType)
	part, err := w.CreatePart(h)
	if err != nil {
		return nil, &IdkollenError{Message: err.Error(), Err: err}
	}
	if _, err := part.Write(data); err != nil {
		return nil, &IdkollenError{Message: err.Error(), Err: err}
	}
	if err := w.Close(); err != nil {
		return nil, &IdkollenError{Message: err.Error(), Err: err}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.client.baseURL+"/v3/document", &buf)
	if err != nil {
		return nil, &IdkollenError{Message: err.Error(), Err: err}
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("User-Agent", "idkollen-client-go/0.1.0")
	e.client.setAuth(req)

	var out DocumentUploadResponse
	if err := e.client.do(req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (e *DocumentEndpoint) Download(ctx context.Context, id string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, e.client.baseURL+"/v3/document/"+id, nil)
	if err != nil {
		return nil, &IdkollenError{Message: err.Error(), Err: err}
	}
	req.Header.Set("User-Agent", "idkollen-client-go/0.1.0")
	e.client.setAuth(req)

	return e.client.send(req)
}

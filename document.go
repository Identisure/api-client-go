package idkollen

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
)

type DocumentUploadResponse struct {
	Id   string `json:"id"`
	Hash string `json:"hash"`
}

type DocumentEndpoint struct {
	client *IdkollenClient
}

func (e *DocumentEndpoint) Upload(ctx context.Context, data []byte, filename string, mimeType string) (*DocumentUploadResponse, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="file"; filename="`+filename+`"`)
	h.Set("Content-Type", mimeType)
	part, err := w.CreatePart(h)
	if err != nil {
		return nil, &IdkollenError{Message: err.Error()}
	}
	if _, err := part.Write(data); err != nil {
		return nil, &IdkollenError{Message: err.Error()}
	}
	if err := w.Close(); err != nil {
		return nil, &IdkollenError{Message: err.Error()}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.client.baseURL+"/document", &buf)
	if err != nil {
		return nil, &IdkollenError{Message: err.Error()}
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("User-Agent", "idkollen-client-go/0.1.0")
	req.SetBasicAuth(e.client.clientID, e.client.clientSecret)

	resp, err := e.client.httpClient.Do(req)
	if err != nil {
		return nil, &IdkollenError{Message: err.Error()}
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &IdkollenError{Message: err.Error()}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiErr IdkollenError
		if jsonErr := json.Unmarshal(body, &apiErr); jsonErr != nil {
			return nil, &IdkollenError{StatusCode: resp.StatusCode, Message: string(body)}
		}
		apiErr.StatusCode = resp.StatusCode
		return nil, &apiErr
	}
	var out DocumentUploadResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, &IdkollenError{Message: err.Error()}
	}
	return &out, nil
}

func (e *DocumentEndpoint) Download(ctx context.Context, id string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, e.client.baseURL+"/document/"+id, nil)
	if err != nil {
		return nil, &IdkollenError{Message: err.Error()}
	}
	req.Header.Set("User-Agent", "idkollen-client-go/0.1.0")
	req.SetBasicAuth(e.client.clientID, e.client.clientSecret)

	resp, err := e.client.httpClient.Do(req)
	if err != nil {
		return nil, &IdkollenError{Message: err.Error()}
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &IdkollenError{Message: err.Error()}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiErr IdkollenError
		if jsonErr := json.Unmarshal(body, &apiErr); jsonErr != nil {
			return nil, &IdkollenError{StatusCode: resp.StatusCode, Message: string(body)}
		}
		apiErr.StatusCode = resp.StatusCode
		return nil, &apiErr
	}
	return body, nil
}

func (e *DocumentEndpoint) Delete(ctx context.Context, id string) error {
	return e.client.delete(ctx, "/document/"+id)
}

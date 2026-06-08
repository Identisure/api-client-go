package idkollen_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	idkollen "github.com/idkollen/idk-clients/go/idkollen"
)

func TestDocumentUpload(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/document" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
			t.Errorf("expected multipart/form-data, got %s", r.Header.Get("Content-Type"))
		}
		_, _ = w.Write([]byte(`{"id":"doc1","hash":"sha256hash"}`))
	}))
	defer srv.Close()

	c := idkollen.NewClientBuilder("id", "secret").BaseURL(srv.URL).Build()
	resp, err := c.Document().Upload(context.Background(), []byte("PDF content"), "test.pdf", "application/pdf")
	if err != nil {
		t.Fatal(err)
	}
	if resp.Id != "doc1" {
		t.Errorf("expected id doc1, got %s", resp.Id)
	}
	if resp.Hash != "sha256hash" {
		t.Errorf("expected hash sha256hash, got %s", resp.Hash)
	}
}

func TestDocumentDownload(t *testing.T) {
	pdfBytes := []byte("PDF content")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/document/doc1" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write(pdfBytes)
	}))
	defer srv.Close()

	c := idkollen.NewClientBuilder("id", "secret").BaseURL(srv.URL).Build()
	data, err := c.Document().Download(context.Background(), "doc1")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, pdfBytes) {
		t.Errorf("unexpected download content")
	}
}

func TestDocumentDelete(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/document/doc1" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := idkollen.NewClientBuilder("id", "secret").BaseURL(srv.URL).Build()
	if err := c.Document().Delete(context.Background(), "doc1"); err != nil {
		t.Fatal(err)
	}
}

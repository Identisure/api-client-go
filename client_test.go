package idkollen_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	idkollen "github.com/idkollen/idk-clients/go/idkollen"
)

func TestBuildDefaults(t *testing.T) {
	c := idkollen.NewClientBuilder("id", "secret").Build()
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestErrorOnNon2xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"unauthorized"}`))
	}))
	defer srv.Close()

	c := idkollen.NewClientBuilder("bad", "creds").BaseURL(srv.URL).Build()
	_, err := c.BankIdSe().AuthStatus(context.Background(), "abc")
	if err == nil {
		t.Fatal("expected error")
	}
	idkErr, ok := err.(*idkollen.IdkollenError)
	if !ok {
		t.Fatalf("expected *IdkollenError, got %T", err)
	}
	if idkErr.StatusCode != 401 {
		t.Errorf("expected 401, got %d", idkErr.StatusCode)
	}
}

func TestRequestTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(2 * time.Second):
		}
	}))
	defer srv.Close()

	c := idkollen.NewClientBuilder("id", "secret").BaseURL(srv.URL).Timeout(50 * time.Millisecond).Build()
	_, err := c.BankIdSe().AuthStatus(context.Background(), "abc")
	var idkErr *idkollen.IdkollenError
	if !errors.As(err, &idkErr) {
		t.Fatalf("expected *IdkollenError, got %T", err)
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected deadline exceeded, got %v", err)
	}
}

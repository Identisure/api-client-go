package idkollen_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	idkollen "github.com/idkollen/idk-clients/go/idkollen"
)

func TestFtnAuth_Pending(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v3/ftn/auth" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"status":"PENDING","id":"ft1","url":"https://example.com/ftn"}`))
	}))
	defer srv.Close()

	c := idkollen.NewClientBuilder("id", "secret").BaseURL(srv.URL).Build()
	status, err := c.Ftn().Auth(context.Background(), idkollen.FtnAuthRequest{})
	if err != nil {
		t.Fatal(err)
	}
	p, ok := status.(*idkollen.FtnPending)
	if !ok {
		t.Fatalf("expected *FtnPending, got %T", status)
	}
	if p.Url != "https://example.com/ftn" {
		t.Errorf("expected url, got %s", p.Url)
	}
}

func TestFtnAgeVerification(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v3/ftn/age-verification" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"status":"PENDING","id":"ft2","url":"https://example.com/age"}`))
	}))
	defer srv.Close()

	c := idkollen.NewClientBuilder("id", "secret").BaseURL(srv.URL).Build()
	minAge := 18
	_, err := c.Ftn().AgeVerification(context.Background(), idkollen.AgeVerificationRequest{MinAge: &minAge})
	if err != nil {
		t.Fatal(err)
	}
}

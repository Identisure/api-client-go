package identisure_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	identisure "github.com/identisure/idk-clients/go/identisure"
)

func TestFtnAuth_Pending(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v3/ftn/auth" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"status":"PENDING","id":"ft1","url":"https://example.com/ftn"}`))
	}))
	defer srv.Close()

	c := identisure.NewClientBuilder("id", "secret").BaseURL(srv.URL).Build()
	status, err := c.Ftn().Auth(context.Background(), identisure.FtnAuthRequest{})
	if err != nil {
		t.Fatal(err)
	}
	p, ok := status.(*identisure.FtnPending)
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

	c := identisure.NewClientBuilder("id", "secret").BaseURL(srv.URL).Build()
	minAge := 18
	_, err := c.Ftn().AgeVerification(context.Background(), identisure.AgeVerificationRequest{MinAge: &minAge})
	if err != nil {
		t.Fatal(err)
	}
}

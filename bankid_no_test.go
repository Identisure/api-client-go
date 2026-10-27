package identisure_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	identisure "github.com/identisure/api-client-go"
)

func TestBankIdNoAuth_Pending(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v3/bankid-no/auth" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"status":"PENDING","id":"no1","url":"https://example.com/auth"}`))
	}))
	defer srv.Close()

	c := identisure.NewClientBuilder("id", "secret").BaseURL(srv.URL).Build()
	status, err := c.BankIdNo().Auth(context.Background(), identisure.BankIdNoAuthRequest{})
	if err != nil {
		t.Fatal(err)
	}
	p, ok := status.(*identisure.BankIdNoPending)
	if !ok {
		t.Fatalf("expected *BankIdNoPending, got %T", status)
	}
	if p.Id != "no1" {
		t.Errorf("expected id no1, got %s", p.Id)
	}
}

func TestBankIdNoSign_Completed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"status":"COMPLETED","id":"no2","ssn":"12345678901","name":"Test","givenName":"Test","surname":"User"}`))
	}))
	defer srv.Close()

	c := identisure.NewClientBuilder("id", "secret").BaseURL(srv.URL).Build()
	status, err := c.BankIdNo().Sign(context.Background(), identisure.BankIdNoSignRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := status.(*identisure.BankIdNoCompleted); !ok {
		t.Fatalf("expected *BankIdNoCompleted, got %T", status)
	}
}

func TestBankIdNoBackchannelAuth(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v3/bankid-no/backchannel/auth" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"status":"PENDING","id":"no3"}`))
	}))
	defer srv.Close()

	c := identisure.NewClientBuilder("id", "secret").BaseURL(srv.URL).Build()
	_, err := c.BankIdNo().BackchannelAuth(context.Background(), identisure.BankIdNoBackchannelAuthRequest{Ssn: "12345678901"})
	if err != nil {
		t.Fatal(err)
	}
}

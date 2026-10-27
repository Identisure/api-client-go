package identisure_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	identisure "github.com/identisure/idk-clients/go/identisure"
)

func TestBankIdSeAuth_Pending(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v3/bankid-se/auth" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"status":"PENDING","id":"abc","autoStartToken":"tok","qrStartToken":"qt","qrStartSecret":"qs"}`))
	}))
	defer srv.Close()

	c := identisure.NewClientBuilder("id", "secret").BaseURL(srv.URL).Build()
	status, err := c.BankIdSe().Auth(context.Background(), identisure.BankIdSeAuthRequest{})
	if err != nil {
		t.Fatal(err)
	}
	p, ok := status.(*identisure.BankIdSePending)
	if !ok {
		t.Fatalf("expected *BankIdSePending, got %T", status)
	}
	if p.Id != "abc" {
		t.Errorf("expected id abc, got %s", p.Id)
	}
}

func TestBankIdSeAuth_Completed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"status":"COMPLETED","id":"abc","ssn":"190001011234","name":"Test User","givenName":"Test","surname":"User"}`))
	}))
	defer srv.Close()

	c := identisure.NewClientBuilder("id", "secret").BaseURL(srv.URL).Build()
	status, err := c.BankIdSe().Auth(context.Background(), identisure.BankIdSeAuthRequest{})
	if err != nil {
		t.Fatal(err)
	}
	comp, ok := status.(*identisure.BankIdSeCompleted)
	if !ok {
		t.Fatalf("expected *BankIdSeCompleted, got %T", status)
	}
	if comp.Ssn != "190001011234" {
		t.Errorf("expected ssn, got %s", comp.Ssn)
	}
}

func TestBankIdSePhoneAuth_PendingPhone(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v3/bankid-se/phone/auth" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"status":"PENDING","id":"phone1","hintCode":"outstandingTransaction"}`))
	}))
	defer srv.Close()

	c := identisure.NewClientBuilder("id", "secret").BaseURL(srv.URL).Build()
	status, err := c.BankIdSe().PhoneAuth(context.Background(), identisure.BankIdSePhoneAuthRequest{Ssn: "190001011234", CallInitiator: "USER"})
	if err != nil {
		t.Fatal(err)
	}
	p, ok := status.(*identisure.BankIdSePendingPhone)
	if !ok {
		t.Fatalf("expected *BankIdSePendingPhone, got %T", status)
	}
	if p.Id != "phone1" {
		t.Errorf("expected id phone1, got %s", p.Id)
	}
}

func TestBankIdSeVerify(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["qrCode"] == nil {
			t.Error("expected qrCode in body")
		}
		_, _ = w.Write([]byte(`{"ssn":"190001011234","name":"Test","givenName":"Test","surname":"User"}`))
	}))
	defer srv.Close()

	c := identisure.NewClientBuilder("id", "secret").BaseURL(srv.URL).Build()
	resp, err := c.BankIdSe().Verify(context.Background(), identisure.BankIdSeVerifyRequest{QrCode: "qrdata"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Ssn != "190001011234" {
		t.Errorf("expected ssn, got %s", resp.Ssn)
	}
}

func TestBankIdSeCancelAuth(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := identisure.NewClientBuilder("id", "secret").BaseURL(srv.URL).Build()
	if err := c.BankIdSe().CancelAuth(context.Background(), "abc"); err != nil {
		t.Fatal(err)
	}
}

func TestBankIdSeWaitForAuth_CompletesImmediately(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"status":"COMPLETED","id":"abc","ssn":"190001011234","name":"Test","givenName":"Test","surname":"User"}`))
	}))
	defer srv.Close()

	c := identisure.NewClientBuilder("id", "secret").BaseURL(srv.URL).Build()
	status, err := c.BankIdSe().WaitForAuth(context.Background(), "abc", identisure.PollOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := status.(*identisure.BankIdSeCompleted); !ok {
		t.Fatalf("expected *BankIdSeCompleted, got %T", status)
	}
}

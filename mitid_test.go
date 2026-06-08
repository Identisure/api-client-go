package idkollen_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	idkollen "github.com/idkollen/idk-clients/go/idkollen"
)

func TestMitIdAuth_Pending(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v3/mitid/auth" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"status":"PENDING","id":"mi1","url":"https://example.com/mitid"}`))
	}))
	defer srv.Close()

	c := idkollen.NewClientBuilder("id", "secret").BaseURL(srv.URL).Build()
	status, err := c.MitId().Auth(context.Background(), idkollen.MitIdAuthRequest{})
	if err != nil {
		t.Fatal(err)
	}
	p, ok := status.(*idkollen.MitIdPending)
	if !ok {
		t.Fatalf("expected *MitIdPending, got %T", status)
	}
	if p.Id != "mi1" {
		t.Errorf("expected id mi1, got %s", p.Id)
	}
}

func TestMitIdBackchannelAuth(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v3/mitid/backchannel/auth" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"status":"PENDING","id":"mi2"}`))
	}))
	defer srv.Close()

	c := idkollen.NewClientBuilder("id", "secret").BaseURL(srv.URL).Build()
	_, err := c.MitId().BackchannelAuth(context.Background(), idkollen.MitIdBackchannelAuthRequest{Ssn: "1234567890", BindingMessage: "Login"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestMitIdSign_Completed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"status":"COMPLETED","id":"mi3","ssn":"1234567890","name":"Test","givenName":"Test","surname":"User","signResult":{"checksum":"abc"}}`))
	}))
	defer srv.Close()

	c := idkollen.NewClientBuilder("id", "secret").BaseURL(srv.URL).Build()
	status, err := c.MitId().Sign(context.Background(), idkollen.MitIdSignRequest{Text: "Sign this"})
	if err != nil {
		t.Fatal(err)
	}
	comp, ok := status.(*idkollen.MitIdCompleted)
	if !ok {
		t.Fatalf("expected *MitIdCompleted, got %T", status)
	}
	if comp.SignResult == nil || comp.SignResult.Checksum != "abc" {
		t.Error("expected signResult.checksum abc")
	}
}

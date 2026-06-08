package idkollen_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	idkollen "github.com/idkollen/idk-clients/go/idkollen"
)

func TestFrejaAuth_Pending(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v3/freja/auth" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"status":"PENDING","id":"fr1","autoStartToken":"ast","qrData":"qrdata"}`))
	}))
	defer srv.Close()

	c := idkollen.NewClientBuilder("id", "secret").BaseURL(srv.URL).Build()
	status, err := c.Freja().Auth(context.Background(), idkollen.FrejaAuthRequest{})
	if err != nil {
		t.Fatal(err)
	}
	p, ok := status.(*idkollen.FrejaPending)
	if !ok {
		t.Fatalf("expected *FrejaPending, got %T", status)
	}
	if p.AutoStartToken != "ast" {
		t.Errorf("expected autoStartToken ast, got %s", p.AutoStartToken)
	}
}

func TestFrejaBackchannelSign(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v3/freja/backchannel/sign" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"status":"PENDING","id":"fr2","autoStartToken":"ast2","qrData":"qr2"}`))
	}))
	defer srv.Close()

	c := idkollen.NewClientBuilder("id", "secret").BaseURL(srv.URL).Build()
	_, err := c.Freja().BackchannelSign(context.Background(), idkollen.FrejaBackchannelSignRequest{Ssn: "190001011234", Country: "SE", Text: "Sign this"})
	if err != nil {
		t.Fatal(err)
	}
}

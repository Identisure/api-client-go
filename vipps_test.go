package idkollen_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	idkollen "github.com/idkollen/idk-clients/go/idkollen"
)

func TestVippsAuth_Pending(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v3/vipps/auth" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"status":"PENDING","id":"vp1","url":"https://example.com/vipps"}`))
	}))
	defer srv.Close()

	c := idkollen.NewClientBuilder("id", "secret").BaseURL(srv.URL).Build()
	status, err := c.Vipps().Auth(context.Background(), idkollen.VippsAuthRequest{})
	if err != nil {
		t.Fatal(err)
	}
	p, ok := status.(*idkollen.VippsPending)
	if !ok {
		t.Fatalf("expected *VippsPending, got %T", status)
	}
	if p.Id != "vp1" {
		t.Errorf("expected id vp1, got %s", p.Id)
	}
}

func TestVippsBackchannelAuth(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v3/vipps/backchannel/auth" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"status":"PENDING","id":"vp2"}`))
	}))
	defer srv.Close()

	c := idkollen.NewClientBuilder("id", "secret").BaseURL(srv.URL).Build()
	_, err := c.Vipps().BackchannelAuth(context.Background(), idkollen.VippsBackchannelAuthRequest{Phone: "+4712345678"})
	if err != nil {
		t.Fatal(err)
	}
}

package identisure_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	identisure "github.com/identisure/api-client-go"
)

func TestVippsAuth_Pending(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v3/vipps/auth" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"status":"PENDING","id":"vp1","url":"https://example.com/vipps"}`))
	}))
	defer srv.Close()

	c := identisure.NewClientBuilder("id", "secret").BaseURL(srv.URL).Build()
	status, err := c.Vipps().Auth(context.Background(), identisure.VippsAuthRequest{})
	if err != nil {
		t.Fatal(err)
	}
	p, ok := status.(*identisure.VippsPending)
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

	c := identisure.NewClientBuilder("id", "secret").BaseURL(srv.URL).Build()
	_, err := c.Vipps().BackchannelAuth(context.Background(), identisure.VippsBackchannelAuthRequest{Phone: "+4712345678"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestVippsAuthStatus_Unknown(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"status":"SOMETHING_NEW","id":"vp1"}`))
	}))
	defer srv.Close()

	c := identisure.NewClientBuilder("id", "secret").BaseURL(srv.URL).Build()
	status, err := c.Vipps().AuthStatus(context.Background(), "vp1")
	if err != nil {
		t.Fatal(err)
	}
	u, ok := status.(*identisure.UnknownStatus)
	if !ok {
		t.Fatalf("expected *UnknownStatus, got %T", status)
	}
	if u.Status != "SOMETHING_NEW" {
		t.Errorf("expected status SOMETHING_NEW, got %s", u.Status)
	}
}

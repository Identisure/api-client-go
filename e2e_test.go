package identisure_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	identisure "github.com/identisure/idk-clients/go/identisure"
)

var fastPoll = identisure.PollOptions{Interval: time.Millisecond}

type mockResponse struct {
	status  int
	fixture string
}

type recordedRequest struct {
	method string
	path   string
	auth   bool
}

type fixtureServer struct {
	t         *testing.T
	mu        sync.Mutex
	responses []mockResponse
	requests  []recordedRequest
	srv       *httptest.Server
}

func newFixtureServer(t *testing.T, responses ...mockResponse) *fixtureServer {
	t.Helper()
	fs := &fixtureServer{t: t, responses: responses}
	fs.srv = httptest.NewServer(http.HandlerFunc(fs.handle))
	t.Cleanup(fs.srv.Close)
	return fs
}

func (fs *fixtureServer) handle(w http.ResponseWriter, r *http.Request) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	user, pass, ok := r.BasicAuth()
	fs.requests = append(fs.requests, recordedRequest{
		method: r.Method,
		path:   r.URL.Path,
		auth:   ok && user == "client-id" && pass == "client-secret",
	})
	if len(fs.responses) == 0 {
		fs.t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	res := fs.responses[0]
	fs.responses = fs.responses[1:]
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(res.status)
	_, _ = w.Write(readFixture(fs.t, res.fixture))
}

func (fs *fixtureServer) client() *identisure.IdentisureClient {
	return identisure.NewClientBuilder("client-id", "client-secret").BaseURL(fs.srv.URL).Build()
}

func (fs *fixtureServer) assertRequests(want ...recordedRequest) {
	fs.t.Helper()
	fs.mu.Lock()
	defer fs.mu.Unlock()
	if len(fs.requests) != len(want) {
		fs.t.Fatalf("expected %d requests, got %d: %+v", len(want), len(fs.requests), fs.requests)
	}
	for i, w := range want {
		got := fs.requests[i]
		if got.method != w.method || got.path != w.path {
			fs.t.Errorf("request %d: expected %s %s, got %s %s", i, w.method, w.path, got.method, got.path)
		}
		if !got.auth {
			fs.t.Errorf("request %d: missing or wrong basic auth", i)
		}
	}
	if len(fs.responses) != 0 {
		fs.t.Errorf("%d queued responses were not served", len(fs.responses))
	}
}

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "fixtures", name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return data
}

func fixtureMap(t *testing.T, name string) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(readFixture(t, name), &m); err != nil {
		t.Fatalf("parse fixture %s: %v", name, err)
	}
	return m
}

// assertMatchesFixture re-encodes the decoded value through its json tags and checks
// that every field of the fixture except the status discriminator survived decoding.
func assertMatchesFixture(t *testing.T, got any, name string) {
	t.Helper()
	data, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	for key, want := range fixtureMap(t, name) {
		if key == "status" {
			continue
		}
		if !reflect.DeepEqual(decoded[key], want) {
			t.Errorf("%s: field %s: expected %v, got %v", name, key, want, decoded[key])
		}
	}
}

func get(path string) recordedRequest  { return recordedRequest{method: http.MethodGet, path: path} }
func post(path string) recordedRequest { return recordedRequest{method: http.MethodPost, path: path} }

func created(fixture string) mockResponse { return mockResponse{http.StatusCreated, fixture} }
func ok(fixture string) mockResponse      { return mockResponse{http.StatusOK, fixture} }

type authFlow struct {
	provider string
	path     string
	id       string
	start    func(context.Context, *identisure.IdentisureClient) (any, error)
	wait     func(context.Context, *identisure.IdentisureClient, string) (any, error)
}

var authFlows = []authFlow{
	{
		provider: "bankid_se", path: "bankid-se", id: "bankid-se-session-1",
		start: func(ctx context.Context, c *identisure.IdentisureClient) (any, error) {
			return c.BankIdSe().Auth(ctx, identisure.BankIdSeAuthRequest{})
		},
		wait: func(ctx context.Context, c *identisure.IdentisureClient, id string) (any, error) {
			return c.BankIdSe().WaitForAuth(ctx, id, fastPoll)
		},
	},
	{
		provider: "bankid_no", path: "bankid-no", id: "bankid-no-session-1",
		start: func(ctx context.Context, c *identisure.IdentisureClient) (any, error) {
			return c.BankIdNo().Auth(ctx, identisure.BankIdNoAuthRequest{})
		},
		wait: func(ctx context.Context, c *identisure.IdentisureClient, id string) (any, error) {
			return c.BankIdNo().WaitForAuth(ctx, id, fastPoll)
		},
	},
	{
		provider: "freja", path: "freja", id: "freja-session-1",
		start: func(ctx context.Context, c *identisure.IdentisureClient) (any, error) {
			return c.Freja().Auth(ctx, identisure.FrejaAuthRequest{})
		},
		wait: func(ctx context.Context, c *identisure.IdentisureClient, id string) (any, error) {
			return c.Freja().WaitForAuth(ctx, id, fastPoll)
		},
	},
	{
		provider: "mitid", path: "mitid", id: "mitid-session-1",
		start: func(ctx context.Context, c *identisure.IdentisureClient) (any, error) {
			return c.MitId().Auth(ctx, identisure.MitIdAuthRequest{})
		},
		wait: func(ctx context.Context, c *identisure.IdentisureClient, id string) (any, error) {
			return c.MitId().WaitForAuth(ctx, id, fastPoll)
		},
	},
	{
		provider: "ftn", path: "ftn", id: "ftn-session-1",
		start: func(ctx context.Context, c *identisure.IdentisureClient) (any, error) {
			return c.Ftn().Auth(ctx, identisure.FtnAuthRequest{})
		},
		wait: func(ctx context.Context, c *identisure.IdentisureClient, id string) (any, error) {
			return c.Ftn().WaitForAuth(ctx, id, fastPoll)
		},
	},
	{
		provider: "vipps", path: "vipps", id: "vipps-session-1",
		start: func(ctx context.Context, c *identisure.IdentisureClient) (any, error) {
			return c.Vipps().Auth(ctx, identisure.VippsAuthRequest{})
		},
		wait: func(ctx context.Context, c *identisure.IdentisureClient, id string) (any, error) {
			return c.Vipps().WaitForAuth(ctx, id, fastPoll)
		},
	},
}

func sessionId(t *testing.T, status any) string {
	t.Helper()
	v := reflect.ValueOf(status)
	if v.Kind() != reflect.Pointer || v.Elem().Kind() != reflect.Struct {
		t.Fatalf("unexpected status %T", status)
	}
	id := v.Elem().FieldByName("Id")
	if !id.IsValid() {
		t.Fatalf("status %T has no Id", status)
	}
	return id.String()
}

func typeName(v any) string {
	return reflect.TypeOf(v).Elem().Name()
}

func TestE2EAuth(t *testing.T) {
	for _, flow := range authFlows {
		t.Run(flow.provider, func(t *testing.T) {
			for _, outcome := range []struct {
				fixture string
				want    string
			}{
				{"auth_completed.json", "Completed"},
				{"auth_failed.json", "Failed"},
				{"auth_unknown.json", "UnknownStatus"},
			} {
				t.Run(outcome.want, func(t *testing.T) {
					fs := newFixtureServer(t,
						created(flow.provider+"/auth_pending.json"),
						ok(flow.provider+"/auth_pending.json"),
						ok(flow.provider+"/"+outcome.fixture),
					)
					c := fs.client()
					ctx := context.Background()

					pending, err := flow.start(ctx, c)
					if err != nil {
						t.Fatal(err)
					}
					if !strings.HasSuffix(typeName(pending), "Pending") {
						t.Fatalf("expected pending status, got %T", pending)
					}
					assertMatchesFixture(t, pending, flow.provider+"/auth_pending.json")

					id := sessionId(t, pending)
					result, err := flow.wait(ctx, c, id)
					if err != nil {
						t.Fatal(err)
					}

					statusPath := "/v3/" + flow.path + "/auth/" + id
					fs.assertRequests(post("/v3/"+flow.path+"/auth"), get(statusPath), get(statusPath))

					switch outcome.want {
					case "UnknownStatus":
						u, isUnknown := result.(*identisure.UnknownStatus)
						if !isUnknown {
							t.Fatalf("expected *UnknownStatus, got %T", result)
						}
						if u.Status != "SOMETHING_NEW" {
							t.Errorf("expected status SOMETHING_NEW, got %s", u.Status)
						}
					default:
						if !strings.HasSuffix(typeName(result), outcome.want) {
							t.Fatalf("expected %s status, got %T", outcome.want, result)
						}
						assertMatchesFixture(t, result, flow.provider+"/"+outcome.fixture)
					}
				})
			}
		})
	}
}

func TestE2ESign(t *testing.T) {
	flows := []authFlow{
		{
			provider: "bankid_se", path: "bankid-se",
			start: func(ctx context.Context, c *identisure.IdentisureClient) (any, error) {
				return c.BankIdSe().Sign(ctx, identisure.BankIdSeSignRequest{})
			},
			wait: func(ctx context.Context, c *identisure.IdentisureClient, id string) (any, error) {
				return c.BankIdSe().WaitForSign(ctx, id, fastPoll)
			},
		},
		{
			provider: "bankid_no", path: "bankid-no",
			start: func(ctx context.Context, c *identisure.IdentisureClient) (any, error) {
				return c.BankIdNo().Sign(ctx, identisure.BankIdNoSignRequest{})
			},
			wait: func(ctx context.Context, c *identisure.IdentisureClient, id string) (any, error) {
				return c.BankIdNo().WaitForSign(ctx, id, fastPoll)
			},
		},
		{
			provider: "freja", path: "freja",
			start: func(ctx context.Context, c *identisure.IdentisureClient) (any, error) {
				return c.Freja().Sign(ctx, identisure.FrejaSignRequest{})
			},
			wait: func(ctx context.Context, c *identisure.IdentisureClient, id string) (any, error) {
				return c.Freja().WaitForSign(ctx, id, fastPoll)
			},
		},
		{
			provider: "mitid", path: "mitid",
			start: func(ctx context.Context, c *identisure.IdentisureClient) (any, error) {
				return c.MitId().Sign(ctx, identisure.MitIdSignRequest{})
			},
			wait: func(ctx context.Context, c *identisure.IdentisureClient, id string) (any, error) {
				return c.MitId().WaitForSign(ctx, id, fastPoll)
			},
		},
	}
	for _, flow := range flows {
		t.Run(flow.provider, func(t *testing.T) {
			fs := newFixtureServer(t,
				created(flow.provider+"/auth_pending.json"),
				ok(flow.provider+"/sign_completed.json"),
			)
			c := fs.client()
			ctx := context.Background()

			pending, err := flow.start(ctx, c)
			if err != nil {
				t.Fatal(err)
			}
			id := sessionId(t, pending)
			result, err := flow.wait(ctx, c, id)
			if err != nil {
				t.Fatal(err)
			}

			fs.assertRequests(post("/v3/"+flow.path+"/sign"), get("/v3/"+flow.path+"/sign/"+id))
			if !strings.HasSuffix(typeName(result), "Completed") {
				t.Fatalf("expected completed status, got %T", result)
			}
			assertMatchesFixture(t, result, flow.provider+"/sign_completed.json")
		})
	}
}

func TestE2EBankIdSePhoneAuth(t *testing.T) {
	fs := newFixtureServer(t,
		created("bankid_se/phone_auth_pending.json"),
		ok("bankid_se/auth_completed.json"),
	)
	c := fs.client()
	ctx := context.Background()

	status, err := c.BankIdSe().PhoneAuth(ctx, identisure.BankIdSePhoneAuthRequest{})
	if err != nil {
		t.Fatal(err)
	}
	pending, isPending := status.(*identisure.BankIdSePendingPhone)
	if !isPending {
		t.Fatalf("expected *BankIdSePendingPhone, got %T", status)
	}
	result, err := c.BankIdSe().WaitForAuth(ctx, pending.Id, fastPoll)
	if err != nil {
		t.Fatal(err)
	}

	fs.assertRequests(post("/v3/bankid-se/phone/auth"), get("/v3/bankid-se/auth/"+pending.Id))
	if _, isCompleted := result.(*identisure.BankIdSeCompleted); !isCompleted {
		t.Fatalf("expected *BankIdSeCompleted, got %T", result)
	}
	assertMatchesFixture(t, result, "bankid_se/auth_completed.json")
}

func TestE2EBankIdSeVerify(t *testing.T) {
	fs := newFixtureServer(t, ok("bankid_se/verify.json"))

	res, err := fs.client().BankIdSe().Verify(context.Background(), identisure.BankIdSeVerifyRequest{QrCode: "qr"})
	if err != nil {
		t.Fatal(err)
	}

	fs.assertRequests(post("/v3/bankid-se/verify"))
	assertMatchesFixture(t, res, "bankid_se/verify.json")
}

func TestE2EAgeVerification(t *testing.T) {
	for _, outcome := range []struct {
		fixture string
		check   func(t *testing.T, status identisure.AgeVerificationStatus)
	}{
		{"completed.json", func(t *testing.T, status identisure.AgeVerificationStatus) {
			if _, isCompleted := status.(*identisure.AgeVerificationCompleted); !isCompleted {
				t.Fatalf("expected *AgeVerificationCompleted, got %T", status)
			}
			assertMatchesFixture(t, status, "age_verification/completed.json")
		}},
		{"failed.json", func(t *testing.T, status identisure.AgeVerificationStatus) {
			if _, isFailed := status.(*identisure.AgeVerificationFailed); !isFailed {
				t.Fatalf("expected *AgeVerificationFailed, got %T", status)
			}
			assertMatchesFixture(t, status, "age_verification/failed.json")
		}},
		{"unknown.json", func(t *testing.T, status identisure.AgeVerificationStatus) {
			u, isUnknown := status.(*identisure.UnknownStatus)
			if !isUnknown {
				t.Fatalf("expected *UnknownStatus, got %T", status)
			}
			if u.Status != "SOMETHING_NEW" {
				t.Errorf("expected status SOMETHING_NEW, got %s", u.Status)
			}
		}},
	} {
		t.Run(outcome.fixture, func(t *testing.T) {
			fs := newFixtureServer(t,
				created("age_verification/pending.json"),
				ok("age_verification/pending.json"),
				ok("age_verification/"+outcome.fixture),
			)
			c := fs.client()
			ctx := context.Background()

			status, err := c.BankIdSe().AgeVerification(ctx, identisure.AgeVerificationRequest{})
			if err != nil {
				t.Fatal(err)
			}
			pending, isPending := status.(*identisure.AgeVerificationPending)
			if !isPending {
				t.Fatalf("expected *AgeVerificationPending, got %T", status)
			}
			assertMatchesFixture(t, pending, "age_verification/pending.json")

			result, err := c.BankIdSe().WaitForAgeVerification(ctx, pending.Id, fastPoll)
			if err != nil {
				t.Fatal(err)
			}

			statusPath := "/v3/bankid-se/age-verification/" + pending.Id
			fs.assertRequests(post("/v3/bankid-se/age-verification"), get(statusPath), get(statusPath))
			outcome.check(t, result)
		})
	}
}

func TestE2EDocumentUpload(t *testing.T) {
	fs := newFixtureServer(t, created("document/upload.json"))

	res, err := fs.client().Document().Upload(context.Background(), []byte("%PDF-1.7"), "contract.pdf", "application/pdf")
	if err != nil {
		t.Fatal(err)
	}

	fs.assertRequests(post("/v3/document"))
	assertMatchesFixture(t, res, "document/upload.json")
}

func TestE2EErrorResponse(t *testing.T) {
	fs := newFixtureServer(t, mockResponse{http.StatusBadRequest, "error.json"})

	_, err := fs.client().Ftn().Auth(context.Background(), identisure.FtnAuthRequest{})

	var apiErr *identisure.IdentisureError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *IdentisureError, got %v", err)
	}
	if apiErr.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", apiErr.StatusCode)
	}
	if want := fixtureMap(t, "error.json")["message"]; apiErr.Message != want {
		t.Errorf("expected message %v, got %s", want, apiErr.Message)
	}
	fs.assertRequests(post("/v3/ftn/auth"))
}

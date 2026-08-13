package idkollen

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Environment int

const (
	Production Environment = iota
	Staging
)

func (e Environment) baseURL() string {
	if e == Staging {
		return "https://stgapi.idkollen.se"
	}
	return "https://api.idkollen.se"
}

// Credentials live in a closure so fmt's %+v cannot print the secret.
type basicAuth func(*http.Request)

type IdkollenClient struct {
	setAuth    basicAuth
	baseURL    string
	httpClient *http.Client
	timeout    time.Duration
}

type IdkollenClientBuilder struct {
	setAuth    basicAuth
	baseURL    string
	httpClient *http.Client
	timeout    time.Duration
}

func NewClientBuilder(clientID, clientSecret string) *IdkollenClientBuilder {
	setAuth := func(req *http.Request) { req.SetBasicAuth(clientID, clientSecret) }
	return &IdkollenClientBuilder{setAuth: setAuth, timeout: 30 * time.Second}
}

func (b *IdkollenClientBuilder) Environment(env Environment) *IdkollenClientBuilder {
	b.baseURL = env.baseURL()
	return b
}

func (b *IdkollenClientBuilder) BaseURL(url string) *IdkollenClientBuilder {
	b.baseURL = strings.TrimRight(url, "/")
	return b
}

func (b *IdkollenClientBuilder) HTTPClient(c *http.Client) *IdkollenClientBuilder {
	b.httpClient = c
	return b
}

// Timeout sets the timeout for each HTTP request, including reading the response body.
// It must be positive. Default: 30s.
func (b *IdkollenClientBuilder) Timeout(d time.Duration) *IdkollenClientBuilder {
	b.timeout = d
	return b
}

func (b *IdkollenClientBuilder) Build() *IdkollenClient {
	baseURL := b.baseURL
	if baseURL == "" {
		baseURL = Production.baseURL()
	}
	httpClient := b.httpClient
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	return &IdkollenClient{
		setAuth:    b.setAuth,
		baseURL:    baseURL,
		httpClient: httpClient,
		timeout:    b.timeout,
	}
}

func (c *IdkollenClient) BankIdSe() *BankIdSeEndpoint { return &BankIdSeEndpoint{client: c} }
func (c *IdkollenClient) BankIdNo() *BankIdNoEndpoint { return &BankIdNoEndpoint{client: c} }
func (c *IdkollenClient) Freja() *FrejaEndpoint       { return &FrejaEndpoint{client: c} }
func (c *IdkollenClient) MitId() *MitIdEndpoint       { return &MitIdEndpoint{client: c} }
func (c *IdkollenClient) Ftn() *FtnEndpoint           { return &FtnEndpoint{client: c} }
func (c *IdkollenClient) Vipps() *VippsEndpoint       { return &VippsEndpoint{client: c} }
func (c *IdkollenClient) Document() *DocumentEndpoint { return &DocumentEndpoint{client: c} }

// IdkollenError is returned on non-2xx responses or network failure.
type IdkollenError struct {
	StatusCode int    `json:"-"`
	Message    string `json:"message"`
	// Err is the underlying network or decoding error, if any.
	Err error `json:"-"`
}

func (e *IdkollenError) Error() string {
	return fmt.Sprintf("idkollen: %d %s", e.StatusCode, e.Message)
}

func (e *IdkollenError) Unwrap() error { return e.Err }

// WaitError is returned only by WaitFor* polling helpers.
type WaitError struct {
	Timeout bool
	Err     error
}

func (e *WaitError) Error() string {
	if e.Timeout {
		return "idkollen: poll timed out"
	}
	return fmt.Sprintf("idkollen: poll error: %v", e.Err)
}

func (e *WaitError) Unwrap() error { return e.Err }

// PollOptions configures WaitFor* polling helpers.
type PollOptions struct {
	Interval time.Duration // default: 2s when zero
}

func (c *IdkollenClient) post(ctx context.Context, path string, body any, out any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return &IdkollenError{Message: err.Error(), Err: err}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return &IdkollenError{Message: err.Error(), Err: err}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "idkollen-client-go/0.1.0")
	c.setAuth(req)
	return c.do(req, out)
}

func (c *IdkollenClient) get(ctx context.Context, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return &IdkollenError{Message: err.Error(), Err: err}
	}
	req.Header.Set("User-Agent", "idkollen-client-go/0.1.0")
	c.setAuth(req)
	return c.do(req, out)
}

func (c *IdkollenClient) delete(ctx context.Context, path string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.baseURL+path, nil)
	if err != nil {
		return &IdkollenError{Message: err.Error(), Err: err}
	}
	req.Header.Set("User-Agent", "idkollen-client-go/0.1.0")
	c.setAuth(req)
	return c.do(req, nil)
}

func (c *IdkollenClient) do(req *http.Request, out any) error {
	body, err := c.send(req)
	if err != nil {
		return err
	}
	if out != nil {
		if err := json.Unmarshal(body, out); err != nil {
			return &IdkollenError{Message: err.Error(), Err: err}
		}
	}
	return nil
}

// send performs req within the client timeout and returns the body of a 2xx response.
func (c *IdkollenClient) send(req *http.Request) ([]byte, error) {
	ctx, cancel := context.WithTimeout(req.Context(), c.timeout)
	defer cancel()
	resp, err := c.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, &IdkollenError{Message: err.Error(), Err: err}
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &IdkollenError{Message: err.Error(), Err: err}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiErr IdkollenError
		if jsonErr := json.Unmarshal(body, &apiErr); jsonErr != nil {
			return nil, &IdkollenError{StatusCode: resp.StatusCode, Message: string(body)}
		}
		apiErr.StatusCode = resp.StatusCode
		return nil, &apiErr
	}
	return body, nil
}

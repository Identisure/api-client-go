package idkollen

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

type IdkollenClient struct {
	clientID     string
	clientSecret string
	baseURL      string
	httpClient   *http.Client
}

type IdkollenClientBuilder struct {
	clientID     string
	clientSecret string
	baseURL      string
	httpClient   *http.Client
}

func NewClientBuilder(clientID, clientSecret string) *IdkollenClientBuilder {
	return &IdkollenClientBuilder{clientID: clientID, clientSecret: clientSecret}
}

func (b *IdkollenClientBuilder) Environment(env Environment) *IdkollenClientBuilder {
	b.baseURL = env.baseURL()
	return b
}

func (b *IdkollenClientBuilder) BaseURL(url string) *IdkollenClientBuilder {
	b.baseURL = url
	return b
}

func (b *IdkollenClientBuilder) HTTPClient(c *http.Client) *IdkollenClientBuilder {
	b.httpClient = c
	return b
}

func (b *IdkollenClientBuilder) Build() *IdkollenClient {
	baseURL := b.baseURL
	if baseURL == "" {
		baseURL = Production.baseURL()
	}
	httpClient := b.httpClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &IdkollenClient{
		clientID:     b.clientID,
		clientSecret: b.clientSecret,
		baseURL:      baseURL,
		httpClient:   httpClient,
	}
}

func (c *IdkollenClient) BankIdSe() *BankIdSeEndpoint  { return &BankIdSeEndpoint{client: c} }
func (c *IdkollenClient) BankIdNo() *BankIdNoEndpoint  { return &BankIdNoEndpoint{client: c} }
func (c *IdkollenClient) Freja() *FrejaEndpoint        { return &FrejaEndpoint{client: c} }
func (c *IdkollenClient) MitId() *MitIdEndpoint        { return &MitIdEndpoint{client: c} }
func (c *IdkollenClient) Ftn() *FtnEndpoint            { return &FtnEndpoint{client: c} }
func (c *IdkollenClient) Vipps() *VippsEndpoint        { return &VippsEndpoint{client: c} }
func (c *IdkollenClient) Document() *DocumentEndpoint  { return &DocumentEndpoint{client: c} }

// IdkollenError is returned on non-2xx responses or network failure.
type IdkollenError struct {
	StatusCode int    `json:"-"`
	Message    string `json:"message"`
}

func (e *IdkollenError) Error() string {
	return fmt.Sprintf("idkollen: %d %s", e.StatusCode, e.Message)
}

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
		return &IdkollenError{Message: err.Error()}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return &IdkollenError{Message: err.Error()}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "idkollen-client-go/0.1.0")
	req.SetBasicAuth(c.clientID, c.clientSecret)
	return c.do(req, out)
}

func (c *IdkollenClient) get(ctx context.Context, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return &IdkollenError{Message: err.Error()}
	}
	req.Header.Set("User-Agent", "idkollen-client-go/0.1.0")
	req.SetBasicAuth(c.clientID, c.clientSecret)
	return c.do(req, out)
}

func (c *IdkollenClient) delete(ctx context.Context, path string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.baseURL+path, nil)
	if err != nil {
		return &IdkollenError{Message: err.Error()}
	}
	req.Header.Set("User-Agent", "idkollen-client-go/0.1.0")
	req.SetBasicAuth(c.clientID, c.clientSecret)
	return c.do(req, nil)
}

func (c *IdkollenClient) do(req *http.Request, out any) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &IdkollenError{Message: err.Error()}
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &IdkollenError{Message: err.Error()}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiErr IdkollenError
		if jsonErr := json.Unmarshal(body, &apiErr); jsonErr != nil {
			return &IdkollenError{StatusCode: resp.StatusCode, Message: string(body)}
		}
		apiErr.StatusCode = resp.StatusCode
		return &apiErr
	}
	if out != nil {
		if err := json.Unmarshal(body, out); err != nil {
			return &IdkollenError{Message: err.Error()}
		}
	}
	return nil
}

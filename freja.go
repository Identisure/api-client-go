package idkollen

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// --- Freja request types ---

type FrejaAuthRequest struct {
	Ssn                  *string `json:"ssn,omitempty"`
	CallbackUrl          *string `json:"callbackUrl,omitempty"`
	MinRegistrationLevel *string `json:"minRegistrationLevel,omitempty"`
	OrgNumber            *string `json:"orgNumber,omitempty"`
	RequestAddress       *bool   `json:"requestAddress,omitempty"`
	RefId                *string `json:"refId,omitempty"`
}

type FrejaBackchannelAuthRequest struct {
	Ssn                  string  `json:"ssn"`
	Country              string  `json:"country"`
	CallbackUrl          *string `json:"callbackUrl,omitempty"`
	MinRegistrationLevel *string `json:"minRegistrationLevel,omitempty"`
	OrgNumber            *string `json:"orgNumber,omitempty"`
	RequestAddress       *bool   `json:"requestAddress,omitempty"`
	RefId                *string `json:"refId,omitempty"`
}

type FrejaSignRequest struct {
	Text                 string  `json:"text"`
	Ssn                  *string `json:"ssn,omitempty"`
	CallbackUrl          *string `json:"callbackUrl,omitempty"`
	MinRegistrationLevel *string `json:"minRegistrationLevel,omitempty"`
	OrgNumber            *string `json:"orgNumber,omitempty"`
	RequestAddress       *bool   `json:"requestAddress,omitempty"`
	RefId                *string `json:"refId,omitempty"`
}

type FrejaBackchannelSignRequest struct {
	Ssn                  string  `json:"ssn"`
	Country              string  `json:"country"`
	Text                 string  `json:"text"`
	CallbackUrl          *string `json:"callbackUrl,omitempty"`
	MinRegistrationLevel *string `json:"minRegistrationLevel,omitempty"`
	OrgNumber            *string `json:"orgNumber,omitempty"`
	RequestAddress       *bool   `json:"requestAddress,omitempty"`
	RefId                *string `json:"refId,omitempty"`
}

// --- Freja status types ---

type FrejaStatus interface {
	isFrejaStatus()
}

type FrejaPending struct {
	Id             string  `json:"id"`
	RefId          *string `json:"refId"`
	AutoStartToken string  `json:"autoStartToken"`
	QrData         string  `json:"qrData"`
}

type FrejaCompleted struct {
	Id                   string  `json:"id"`
	RefId                *string `json:"refId"`
	Ssn                  string  `json:"ssn"`
	Country              string  `json:"country"`
	Name                 string  `json:"name"`
	GivenName            string  `json:"givenName"`
	Surname              string  `json:"surname"`
	Address              *string `json:"address"`
	CompanySignatoryText *string `json:"companySignatoryText"`
}

type FrejaFailed struct {
	Id    string  `json:"id"`
	RefId *string `json:"refId"`
	Error string  `json:"error"`
}

func (*FrejaPending) isFrejaStatus()   {}
func (*FrejaCompleted) isFrejaStatus() {}
func (*FrejaFailed) isFrejaStatus()    {}

type frejaStatusEnvelope struct {
	Status string `json:"status"`
}

func unmarshalFrejaStatus(data []byte) (FrejaStatus, error) {
	var env frejaStatusEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, err
	}
	switch env.Status {
	case "PENDING":
		var v FrejaPending
		if err := json.Unmarshal(data, &v); err != nil {
			return nil, err
		}
		return &v, nil
	case "COMPLETED":
		var v FrejaCompleted
		if err := json.Unmarshal(data, &v); err != nil {
			return nil, err
		}
		return &v, nil
	case "FAILED":
		var v FrejaFailed
		if err := json.Unmarshal(data, &v); err != nil {
			return nil, err
		}
		return &v, nil
	default:
		return nil, fmt.Errorf("unknown freja status: %s", env.Status)
	}
}

// --- Freja endpoint ---

type FrejaEndpoint struct {
	client *IdkollenClient
}

func (e *FrejaEndpoint) Auth(ctx context.Context, req FrejaAuthRequest) (FrejaStatus, error) {
	var raw json.RawMessage
	if err := e.client.post(ctx, "/v3/freja/auth", req, &raw); err != nil {
		return nil, err
	}
	return unmarshalFrejaStatus(raw)
}

func (e *FrejaEndpoint) BackchannelAuth(ctx context.Context, req FrejaBackchannelAuthRequest) (FrejaStatus, error) {
	var raw json.RawMessage
	if err := e.client.post(ctx, "/v3/freja/backchannel/auth", req, &raw); err != nil {
		return nil, err
	}
	return unmarshalFrejaStatus(raw)
}

func (e *FrejaEndpoint) Sign(ctx context.Context, req FrejaSignRequest) (FrejaStatus, error) {
	var raw json.RawMessage
	if err := e.client.post(ctx, "/v3/freja/sign", req, &raw); err != nil {
		return nil, err
	}
	return unmarshalFrejaStatus(raw)
}

func (e *FrejaEndpoint) BackchannelSign(ctx context.Context, req FrejaBackchannelSignRequest) (FrejaStatus, error) {
	var raw json.RawMessage
	if err := e.client.post(ctx, "/v3/freja/backchannel/sign", req, &raw); err != nil {
		return nil, err
	}
	return unmarshalFrejaStatus(raw)
}

func (e *FrejaEndpoint) AuthStatus(ctx context.Context, id string) (FrejaStatus, error) {
	var raw json.RawMessage
	if err := e.client.get(ctx, "/v3/freja/auth/"+id, &raw); err != nil {
		return nil, err
	}
	return unmarshalFrejaStatus(raw)
}

func (e *FrejaEndpoint) SignStatus(ctx context.Context, id string) (FrejaStatus, error) {
	var raw json.RawMessage
	if err := e.client.get(ctx, "/v3/freja/sign/"+id, &raw); err != nil {
		return nil, err
	}
	return unmarshalFrejaStatus(raw)
}

func (e *FrejaEndpoint) CancelAuth(ctx context.Context, id string) error {
	return e.client.delete(ctx, "/v3/freja/auth/"+id)
}

func (e *FrejaEndpoint) CancelSign(ctx context.Context, id string) error {
	return e.client.delete(ctx, "/v3/freja/sign/"+id)
}

func (e *FrejaEndpoint) WaitForAuth(ctx context.Context, id string, opts PollOptions) (FrejaStatus, error) {
	return pollFreja(ctx, opts, func() (FrejaStatus, error) { return e.AuthStatus(ctx, id) })
}

func (e *FrejaEndpoint) WaitForSign(ctx context.Context, id string, opts PollOptions) (FrejaStatus, error) {
	return pollFreja(ctx, opts, func() (FrejaStatus, error) { return e.SignStatus(ctx, id) })
}

func pollFreja(ctx context.Context, opts PollOptions, fn func() (FrejaStatus, error)) (FrejaStatus, error) {
	interval := opts.Interval
	if interval == 0 {
		interval = 2 * time.Second
	}
	for {
		status, err := fn()
		if err != nil {
			return nil, &WaitError{Err: err}
		}
		if _, ok := status.(*FrejaPending); !ok {
			return status, nil
		}
		select {
		case <-ctx.Done():
			return nil, &WaitError{Timeout: true}
		case <-time.After(interval):
		}
	}
}

package idkollen

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// --- MitID request types ---

type MitIdAuthRequest struct {
	RedirectUrl    *string `json:"redirectUrl,omitempty"`
	ReferenceText  *string `json:"referenceText,omitempty"`
	RequestPhone   *bool   `json:"requestPhone,omitempty"`
	RequestEmail   *bool   `json:"requestEmail,omitempty"`
	RequestAddress *bool   `json:"requestAddress,omitempty"`
	RefId          *string `json:"refId,omitempty"`
}

type MitIdBackchannelAuthRequest struct {
	Ssn            string  `json:"ssn"`
	BindingMessage string  `json:"bindingMessage"`
	CallbackUrl    *string `json:"callbackUrl,omitempty"`
	RefId          *string `json:"refId,omitempty"`
}

type MitIdSignRequest struct {
	Text        string  `json:"text"`
	RedirectUrl *string `json:"redirectUrl,omitempty"`
	RefId       *string `json:"refId,omitempty"`
}

// --- MitID status types ---

type MitIdStatus interface {
	isMitIdStatus()
}

type MitIdPending struct {
	Id             string  `json:"id"`
	RefId          *string `json:"refId"`
	Url            *string `json:"url"`
	BindingMessage *string `json:"bindingMessage"`
}

type MitIdSignResult struct {
	Checksum string `json:"checksum"`
}

type MitIdCompleted struct {
	Id         string           `json:"id"`
	RefId      *string          `json:"refId"`
	Ssn        string           `json:"ssn"`
	Name       string           `json:"name"`
	GivenName  string           `json:"givenName"`
	Surname    string           `json:"surname"`
	Phone      *string          `json:"phone"`
	Email      *string          `json:"email"`
	Address    *string          `json:"address"`
	BirthDate  *string          `json:"birthDate"`
	Pid        *string          `json:"pid"`
	BankId     *string          `json:"bankId"`
	SignResult *MitIdSignResult `json:"signResult"`
}

type MitIdFailed struct {
	Id    string  `json:"id"`
	RefId *string `json:"refId"`
	Error string  `json:"error"`
}

func (*MitIdPending) isMitIdStatus()   {}
func (*MitIdCompleted) isMitIdStatus() {}
func (*MitIdFailed) isMitIdStatus()    {}

type mitIdStatusEnvelope struct {
	Status string `json:"status"`
}

func unmarshalMitIdStatus(data []byte) (MitIdStatus, error) {
	var env mitIdStatusEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, err
	}
	switch env.Status {
	case "PENDING":
		var v MitIdPending
		if err := json.Unmarshal(data, &v); err != nil {
			return nil, err
		}
		return &v, nil
	case "COMPLETED":
		var v MitIdCompleted
		if err := json.Unmarshal(data, &v); err != nil {
			return nil, err
		}
		return &v, nil
	case "FAILED":
		var v MitIdFailed
		if err := json.Unmarshal(data, &v); err != nil {
			return nil, err
		}
		return &v, nil
	default:
		return nil, fmt.Errorf("unknown mitid status: %s", env.Status)
	}
}

// --- MitID endpoint ---

type MitIdEndpoint struct {
	client *IdkollenClient
}

func (e *MitIdEndpoint) Auth(ctx context.Context, req MitIdAuthRequest) (MitIdStatus, error) {
	var raw json.RawMessage
	if err := e.client.post(ctx, "/v3/mitid/auth", req, &raw); err != nil {
		return nil, err
	}
	return unmarshalMitIdStatus(raw)
}

func (e *MitIdEndpoint) BackchannelAuth(ctx context.Context, req MitIdBackchannelAuthRequest) (MitIdStatus, error) {
	var raw json.RawMessage
	if err := e.client.post(ctx, "/v3/mitid/backchannel/auth", req, &raw); err != nil {
		return nil, err
	}
	return unmarshalMitIdStatus(raw)
}

func (e *MitIdEndpoint) Sign(ctx context.Context, req MitIdSignRequest) (MitIdStatus, error) {
	var raw json.RawMessage
	if err := e.client.post(ctx, "/v3/mitid/sign", req, &raw); err != nil {
		return nil, err
	}
	return unmarshalMitIdStatus(raw)
}

func (e *MitIdEndpoint) AuthStatus(ctx context.Context, id string) (MitIdStatus, error) {
	var raw json.RawMessage
	if err := e.client.get(ctx, "/v3/mitid/auth/"+id, &raw); err != nil {
		return nil, err
	}
	return unmarshalMitIdStatus(raw)
}

func (e *MitIdEndpoint) SignStatus(ctx context.Context, id string) (MitIdStatus, error) {
	var raw json.RawMessage
	if err := e.client.get(ctx, "/v3/mitid/sign/"+id, &raw); err != nil {
		return nil, err
	}
	return unmarshalMitIdStatus(raw)
}

func (e *MitIdEndpoint) CancelAuth(ctx context.Context, id string) error {
	return e.client.delete(ctx, "/v3/mitid/auth/"+id)
}

func (e *MitIdEndpoint) CancelSign(ctx context.Context, id string) error {
	return e.client.delete(ctx, "/v3/mitid/sign/"+id)
}

func (e *MitIdEndpoint) WaitForAuth(ctx context.Context, id string, opts PollOptions) (MitIdStatus, error) {
	return pollMitId(ctx, opts, func() (MitIdStatus, error) { return e.AuthStatus(ctx, id) })
}

func (e *MitIdEndpoint) WaitForSign(ctx context.Context, id string, opts PollOptions) (MitIdStatus, error) {
	return pollMitId(ctx, opts, func() (MitIdStatus, error) { return e.SignStatus(ctx, id) })
}

func pollMitId(ctx context.Context, opts PollOptions, fn func() (MitIdStatus, error)) (MitIdStatus, error) {
	interval := opts.Interval
	if interval == 0 {
		interval = 2 * time.Second
	}
	for {
		status, err := fn()
		if err != nil {
			return nil, &WaitError{Err: err}
		}
		if _, ok := status.(*MitIdPending); !ok {
			return status, nil
		}
		select {
		case <-ctx.Done():
			return nil, &WaitError{Timeout: true}
		case <-time.After(interval):
		}
	}
}

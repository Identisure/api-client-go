package idkollen

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// --- FTN request types ---

type FtnAuthRequest struct {
	RedirectUrl    *string `json:"redirectUrl,omitempty"`
	RequestPhone   *bool   `json:"requestPhone,omitempty"`
	RequestEmail   *bool   `json:"requestEmail,omitempty"`
	RequestAddress *bool   `json:"requestAddress,omitempty"`
	RefId          *string `json:"refId,omitempty"`
}

// --- FTN status types ---

type FtnStatus interface {
	isFtnStatus()
}

type FtnPending struct {
	Id    string  `json:"id"`
	RefId *string `json:"refId"`
	Url   string  `json:"url"`
}

type FtnCompleted struct {
	Id        string  `json:"id"`
	RefId     *string `json:"refId"`
	Ssn       string  `json:"ssn"`
	Name      string  `json:"name"`
	GivenName string  `json:"givenName"`
	Surname   string  `json:"surname"`
	Phone     *string `json:"phone"`
	Email     *string `json:"email"`
	Address   *string `json:"address"`
	BirthDate *string `json:"birthDate"`
	Pid       *string `json:"pid"`
	BankId    *string `json:"bankId"`
}

type FtnFailed struct {
	Id    string  `json:"id"`
	RefId *string `json:"refId"`
	Error string  `json:"error"`
}

func (*FtnPending) isFtnStatus()   {}
func (*FtnCompleted) isFtnStatus() {}
func (*FtnFailed) isFtnStatus()    {}

type ftnStatusEnvelope struct {
	Status string `json:"status"`
}

func unmarshalFtnStatus(data []byte) (FtnStatus, error) {
	var env ftnStatusEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, err
	}
	switch env.Status {
	case "PENDING":
		var v FtnPending
		if err := json.Unmarshal(data, &v); err != nil {
			return nil, err
		}
		return &v, nil
	case "COMPLETED":
		var v FtnCompleted
		if err := json.Unmarshal(data, &v); err != nil {
			return nil, err
		}
		return &v, nil
	case "FAILED":
		var v FtnFailed
		if err := json.Unmarshal(data, &v); err != nil {
			return nil, err
		}
		return &v, nil
	default:
		return nil, fmt.Errorf("unknown ftn status: %s", env.Status)
	}
}

// --- FTN endpoint ---

type FtnEndpoint struct {
	client *IdkollenClient
}

func (e *FtnEndpoint) Auth(ctx context.Context, req FtnAuthRequest) (FtnStatus, error) {
	var raw json.RawMessage
	if err := e.client.post(ctx, "/v3/ftn/auth", req, &raw); err != nil {
		return nil, err
	}
	return unmarshalFtnStatus(raw)
}

func (e *FtnEndpoint) AgeVerification(ctx context.Context, req AgeVerificationRequest) (AgeVerificationStatus, error) {
	var raw json.RawMessage
	if err := e.client.post(ctx, "/v3/ftn/age-verification", req, &raw); err != nil {
		return nil, err
	}
	return unmarshalAgeVerificationStatus(raw)
}

func (e *FtnEndpoint) AuthStatus(ctx context.Context, id string) (FtnStatus, error) {
	var raw json.RawMessage
	if err := e.client.get(ctx, "/v3/ftn/auth/"+id, &raw); err != nil {
		return nil, err
	}
	return unmarshalFtnStatus(raw)
}

func (e *FtnEndpoint) AgeVerificationStatus(ctx context.Context, id string) (AgeVerificationStatus, error) {
	var raw json.RawMessage
	if err := e.client.get(ctx, "/v3/ftn/age-verification/"+id, &raw); err != nil {
		return nil, err
	}
	return unmarshalAgeVerificationStatus(raw)
}

func (e *FtnEndpoint) CancelAuth(ctx context.Context, id string) error {
	return e.client.delete(ctx, "/v3/ftn/auth/"+id)
}

func (e *FtnEndpoint) CancelAgeVerification(ctx context.Context, id string) error {
	return e.client.delete(ctx, "/v3/ftn/age-verification/"+id)
}

func (e *FtnEndpoint) WaitForAuth(ctx context.Context, id string, opts PollOptions) (FtnStatus, error) {
	interval := opts.Interval
	if interval == 0 {
		interval = 2 * time.Second
	}
	for {
		status, err := e.AuthStatus(ctx, id)
		if err != nil {
			return nil, &WaitError{Err: err}
		}
		if _, ok := status.(*FtnPending); !ok {
			return status, nil
		}
		select {
		case <-ctx.Done():
			return nil, &WaitError{Timeout: true}
		case <-time.After(interval):
		}
	}
}

func (e *FtnEndpoint) WaitForAgeVerification(ctx context.Context, id string, opts PollOptions) (AgeVerificationStatus, error) {
	interval := opts.Interval
	if interval == 0 {
		interval = 2 * time.Second
	}
	for {
		status, err := e.AgeVerificationStatus(ctx, id)
		if err != nil {
			return nil, &WaitError{Err: err}
		}
		if _, ok := status.(*AgeVerificationPending); !ok {
			return status, nil
		}
		select {
		case <-ctx.Done():
			return nil, &WaitError{Timeout: true}
		case <-time.After(interval):
		}
	}
}

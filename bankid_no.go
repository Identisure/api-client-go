package idkollen

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// --- BankID NO request types ---

type BankIdNoAuthRequest struct {
	RedirectUrl    *string `json:"redirectUrl,omitempty"`
	RequestSsn     *bool   `json:"requestSsn,omitempty"`
	RequestPhone   *bool   `json:"requestPhone,omitempty"`
	RequestEmail   *bool   `json:"requestEmail,omitempty"`
	RequestAddress *bool   `json:"requestAddress,omitempty"`
	RefId          *string `json:"refId,omitempty"`
	AppCallbackUri *string `json:"appCallbackUri,omitempty"`
}

type BankIdNoBackchannelAuthRequest struct {
	Ssn         string  `json:"ssn"`
	CallbackUrl *string `json:"callbackUrl,omitempty"`
	RefId       *string `json:"refId,omitempty"`
}

type BankIdNoSignRequest struct {
	RedirectUrl    *string  `json:"redirectUrl,omitempty"`
	Text           *string  `json:"text,omitempty"`
	Documents      []string `json:"documents,omitempty"`
	RequestSsn     *bool    `json:"requestSsn,omitempty"`
	RequestPhone   *bool    `json:"requestPhone,omitempty"`
	RequestEmail   *bool    `json:"requestEmail,omitempty"`
	RequestAddress *bool    `json:"requestAddress,omitempty"`
	RefId          *string  `json:"refId,omitempty"`
}

// --- BankID NO status types ---

type BankIdNoStatus interface {
	isBankIdNoStatus()
}

type BankIdNoPending struct {
	Id             string  `json:"id"`
	RefId          *string `json:"refId"`
	Url            *string `json:"url"`
	BindingMessage *string `json:"bindingMessage"`
}

type BankIdNoSignResult struct {
	EndUser  string `json:"endUser"`
	Merchant string `json:"merchant"`
	Hash     string `json:"hash"`
}

type BankIdNoSignedDocument struct {
	Id   string `json:"id"`
	Hash string `json:"hash"`
}

type BankIdNoCompleted struct {
	Id              string                   `json:"id"`
	RefId           *string                  `json:"refId"`
	Ssn             string                   `json:"ssn"`
	Name            string                   `json:"name"`
	GivenName       string                   `json:"givenName"`
	Surname         string                   `json:"surname"`
	Phone           *string                  `json:"phone"`
	Email           *string                  `json:"email"`
	Address         *string                  `json:"address"`
	BirthDate       *string                  `json:"birthDate"`
	Pid             *string                  `json:"pid"`
	BankId          *string                  `json:"bankId"`
	SignResult      *BankIdNoSignResult      `json:"signResult"`
	SignedDocuments []BankIdNoSignedDocument `json:"signedDocuments"`
}

type BankIdNoFailed struct {
	Id    string  `json:"id"`
	RefId *string `json:"refId"`
	Error string  `json:"error"`
}

func (*BankIdNoPending) isBankIdNoStatus()   {}
func (*BankIdNoCompleted) isBankIdNoStatus() {}
func (*BankIdNoFailed) isBankIdNoStatus()    {}

type bankIdNoStatusEnvelope struct {
	Status string `json:"status"`
}

func unmarshalBankIdNoStatus(data []byte) (BankIdNoStatus, error) {
	var env bankIdNoStatusEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, err
	}
	switch env.Status {
	case "PENDING":
		var v BankIdNoPending
		if err := json.Unmarshal(data, &v); err != nil {
			return nil, err
		}
		return &v, nil
	case "COMPLETED":
		var v BankIdNoCompleted
		if err := json.Unmarshal(data, &v); err != nil {
			return nil, err
		}
		return &v, nil
	case "FAILED":
		var v BankIdNoFailed
		if err := json.Unmarshal(data, &v); err != nil {
			return nil, err
		}
		return &v, nil
	default:
		return nil, fmt.Errorf("unknown bankid-no status: %s", env.Status)
	}
}

// --- BankID NO endpoint ---

type BankIdNoEndpoint struct {
	client *IdkollenClient
}

func (e *BankIdNoEndpoint) Auth(ctx context.Context, req BankIdNoAuthRequest) (BankIdNoStatus, error) {
	var raw json.RawMessage
	if err := e.client.post(ctx, "/v3/bankid-no/auth", req, &raw); err != nil {
		return nil, err
	}
	return unmarshalBankIdNoStatus(raw)
}

func (e *BankIdNoEndpoint) BackchannelAuth(ctx context.Context, req BankIdNoBackchannelAuthRequest) (BankIdNoStatus, error) {
	var raw json.RawMessage
	if err := e.client.post(ctx, "/v3/bankid-no/backchannel/auth", req, &raw); err != nil {
		return nil, err
	}
	return unmarshalBankIdNoStatus(raw)
}

func (e *BankIdNoEndpoint) Sign(ctx context.Context, req BankIdNoSignRequest) (BankIdNoStatus, error) {
	var raw json.RawMessage
	if err := e.client.post(ctx, "/v3/bankid-no/sign", req, &raw); err != nil {
		return nil, err
	}
	return unmarshalBankIdNoStatus(raw)
}

func (e *BankIdNoEndpoint) AuthStatus(ctx context.Context, id string) (BankIdNoStatus, error) {
	var raw json.RawMessage
	if err := e.client.get(ctx, "/v3/bankid-no/auth/"+id, &raw); err != nil {
		return nil, err
	}
	return unmarshalBankIdNoStatus(raw)
}

func (e *BankIdNoEndpoint) SignStatus(ctx context.Context, id string) (BankIdNoStatus, error) {
	var raw json.RawMessage
	if err := e.client.get(ctx, "/v3/bankid-no/sign/"+id, &raw); err != nil {
		return nil, err
	}
	return unmarshalBankIdNoStatus(raw)
}

func (e *BankIdNoEndpoint) CancelAuth(ctx context.Context, id string) error {
	return e.client.delete(ctx, "/v3/bankid-no/auth/"+id)
}

func (e *BankIdNoEndpoint) CancelSign(ctx context.Context, id string) error {
	return e.client.delete(ctx, "/v3/bankid-no/sign/"+id)
}

func (e *BankIdNoEndpoint) AgeVerification(ctx context.Context, req AgeVerificationRequest) (AgeVerificationStatus, error) {
	var raw json.RawMessage
	if err := e.client.post(ctx, "/v3/bankid-no/age-verification", req, &raw); err != nil {
		return nil, err
	}
	return unmarshalAgeVerificationStatus(raw)
}

func (e *BankIdNoEndpoint) AgeVerificationStatus(ctx context.Context, id string) (AgeVerificationStatus, error) {
	var raw json.RawMessage
	if err := e.client.get(ctx, "/v3/bankid-no/age-verification/"+id, &raw); err != nil {
		return nil, err
	}
	return unmarshalAgeVerificationStatus(raw)
}

func (e *BankIdNoEndpoint) CancelAgeVerification(ctx context.Context, id string) error {
	return e.client.delete(ctx, "/v3/bankid-no/age-verification/"+id)
}

func (e *BankIdNoEndpoint) WaitForAgeVerification(ctx context.Context, id string, opts PollOptions) (AgeVerificationStatus, error) {
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

func (e *BankIdNoEndpoint) WaitForAuth(ctx context.Context, id string, opts PollOptions) (BankIdNoStatus, error) {
	return pollBankIdNo(ctx, opts, func() (BankIdNoStatus, error) { return e.AuthStatus(ctx, id) })
}

func (e *BankIdNoEndpoint) WaitForSign(ctx context.Context, id string, opts PollOptions) (BankIdNoStatus, error) {
	return pollBankIdNo(ctx, opts, func() (BankIdNoStatus, error) { return e.SignStatus(ctx, id) })
}

func pollBankIdNo(ctx context.Context, opts PollOptions, fn func() (BankIdNoStatus, error)) (BankIdNoStatus, error) {
	interval := opts.Interval
	if interval == 0 {
		interval = 2 * time.Second
	}
	for {
		status, err := fn()
		if err != nil {
			return nil, &WaitError{Err: err}
		}
		if _, ok := status.(*BankIdNoPending); !ok {
			return status, nil
		}
		select {
		case <-ctx.Done():
			return nil, &WaitError{Timeout: true}
		case <-time.After(interval):
		}
	}
}

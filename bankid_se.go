package idkollen

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// --- Shared AgeVerification types (also used by FtnEndpoint) ---

type AgeVerificationRequest struct {
	MinAge      *int    `json:"minAge,omitempty"`
	MaxAge      *int    `json:"maxAge,omitempty"`
	RefId       *string `json:"refId,omitempty"`
	CallbackUrl *string `json:"callbackUrl,omitempty"`
	RedirectUrl *string `json:"redirectUrl,omitempty"`
}

type AgeVerificationStatus interface {
	isAgeVerificationStatus()
}

type AgeVerificationPending struct {
	Id     string  `json:"id"`
	Url    *string `json:"url"`
	MinAge *int    `json:"minAge"`
	MaxAge *int    `json:"maxAge"`
}

type AgeVerificationCompleted struct {
	Id          string `json:"id"`
	AgeVerified bool   `json:"ageVerified"`
}

type AgeVerificationFailed struct {
	Id    string `json:"id"`
	Error string `json:"error"`
}

func (*AgeVerificationPending) isAgeVerificationStatus()   {}
func (*AgeVerificationCompleted) isAgeVerificationStatus() {}
func (*AgeVerificationFailed) isAgeVerificationStatus()    {}

type ageVerificationEnvelope struct {
	Status string `json:"status"`
}

func unmarshalAgeVerificationStatus(data []byte) (AgeVerificationStatus, error) {
	var env ageVerificationEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, err
	}
	switch env.Status {
	case "PENDING":
		var v AgeVerificationPending
		if err := json.Unmarshal(data, &v); err != nil {
			return nil, err
		}
		return &v, nil
	case "COMPLETED":
		var v AgeVerificationCompleted
		if err := json.Unmarshal(data, &v); err != nil {
			return nil, err
		}
		return &v, nil
	case "FAILED":
		var v AgeVerificationFailed
		if err := json.Unmarshal(data, &v); err != nil {
			return nil, err
		}
		return &v, nil
	default:
		return nil, fmt.Errorf("unknown age verification status: %s", env.Status)
	}
}

// --- BankID SE request types ---

type BankIdSeAuthRequest struct {
	Ssn            *string `json:"ssn,omitempty"`
	IpAddress      *string `json:"ipAddress,omitempty"`
	CallbackUrl    *string `json:"callbackUrl,omitempty"`
	PinRequired    *bool   `json:"pinRequired,omitempty"`
	Intent         *string `json:"intent,omitempty"`
	OrgNumber      *string `json:"orgNumber,omitempty"`
	RequestAddress *bool   `json:"requestAddress,omitempty"`
	RefId          *string `json:"refId,omitempty"`
}

type BankIdSePhoneAuthRequest struct {
	Ssn            string  `json:"ssn"`
	CallInitiator  string  `json:"callInitiator"`
	CallbackUrl    *string `json:"callbackUrl,omitempty"`
	PinRequired    *bool   `json:"pinRequired,omitempty"`
	Intent         *string `json:"intent,omitempty"`
	OrgNumber      *string `json:"orgNumber,omitempty"`
	RequestAddress *bool   `json:"requestAddress,omitempty"`
	RefId          *string `json:"refId,omitempty"`
}

type BankIdSeSignRequest struct {
	Text           string  `json:"text"`
	Ssn            *string `json:"ssn,omitempty"`
	IpAddress      *string `json:"ipAddress,omitempty"`
	CallbackUrl    *string `json:"callbackUrl,omitempty"`
	PinRequired    *bool   `json:"pinRequired,omitempty"`
	Digest         *string `json:"digest,omitempty"`
	OrgNumber      *string `json:"orgNumber,omitempty"`
	RequestAddress *bool   `json:"requestAddress,omitempty"`
	RefId          *string `json:"refId,omitempty"`
}

type BankIdSePhoneSignRequest struct {
	Ssn            string  `json:"ssn"`
	CallInitiator  string  `json:"callInitiator"`
	Text           string  `json:"text"`
	CallbackUrl    *string `json:"callbackUrl,omitempty"`
	PinRequired    *bool   `json:"pinRequired,omitempty"`
	Digest         *string `json:"digest,omitempty"`
	OrgNumber      *string `json:"orgNumber,omitempty"`
	RequestAddress *bool   `json:"requestAddress,omitempty"`
	RefId          *string `json:"refId,omitempty"`
}

type BankIdSeVerifyRequest struct {
	QrCode string `json:"qrCode"`
}

// --- BankID SE status types ---

type BankIdSeStatus interface {
	isBankIdSeStatus()
}

type BankIdSePhoneStatus interface {
	isBankIdSePhoneStatus()
}

type BankIdSePending struct {
	Id             string  `json:"id"`
	RefId          *string `json:"refId"`
	AutoStartToken *string `json:"autoStartToken"`
	QrStartToken   *string `json:"qrStartToken"`
	QrStartSecret  *string `json:"qrStartSecret"`
	HintCode       *string `json:"hintCode"`
}

type BankIdSePendingPhone struct {
	Id       string  `json:"id"`
	RefId    *string `json:"refId"`
	HintCode *string `json:"hintCode"`
}

type BankIdSeCompleted struct {
	Id                   string  `json:"id"`
	RefId                *string `json:"refId"`
	Ssn                  string  `json:"ssn"`
	Name                 string  `json:"name"`
	GivenName            string  `json:"givenName"`
	Surname              string  `json:"surname"`
	CertStartDate        *string `json:"certStartDate"`
	Address              *string `json:"address"`
	CompanySignatoryText *string `json:"companySignatoryText"`
}

type BankIdSeFailed struct {
	Id    string  `json:"id"`
	RefId *string `json:"refId"`
	Error string  `json:"error"`
}

func (*BankIdSePending) isBankIdSeStatus()        {}
func (*BankIdSeCompleted) isBankIdSeStatus()      {}
func (*BankIdSeFailed) isBankIdSeStatus()         {}
func (*BankIdSeCompleted) isBankIdSePhoneStatus() {}
func (*BankIdSeFailed) isBankIdSePhoneStatus()    {}
func (*BankIdSePendingPhone) isBankIdSePhoneStatus() {}

type BankIdSeVerifyResponse struct {
	Ssn        string  `json:"ssn"`
	Name       string  `json:"name"`
	GivenName  string  `json:"givenName"`
	Surname    string  `json:"surname"`
	Age        *int    `json:"age"`
	VerifiedAt *string `json:"verifiedAt"`
}

type bankIdSeStatusEnvelope struct {
	Status string `json:"status"`
}

func unmarshalBankIdSeStatus(data []byte) (BankIdSeStatus, error) {
	var env bankIdSeStatusEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, err
	}
	switch env.Status {
	case "PENDING":
		var v BankIdSePending
		if err := json.Unmarshal(data, &v); err != nil {
			return nil, err
		}
		return &v, nil
	case "COMPLETED":
		var v BankIdSeCompleted
		if err := json.Unmarshal(data, &v); err != nil {
			return nil, err
		}
		return &v, nil
	case "FAILED":
		var v BankIdSeFailed
		if err := json.Unmarshal(data, &v); err != nil {
			return nil, err
		}
		return &v, nil
	default:
		return nil, fmt.Errorf("unknown bankid-se status: %s", env.Status)
	}
}

func unmarshalBankIdSePhoneStatus(data []byte) (BankIdSePhoneStatus, error) {
	var env bankIdSeStatusEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, err
	}
	switch env.Status {
	case "PENDING":
		var v BankIdSePendingPhone
		if err := json.Unmarshal(data, &v); err != nil {
			return nil, err
		}
		return &v, nil
	case "COMPLETED":
		var v BankIdSeCompleted
		if err := json.Unmarshal(data, &v); err != nil {
			return nil, err
		}
		return &v, nil
	case "FAILED":
		var v BankIdSeFailed
		if err := json.Unmarshal(data, &v); err != nil {
			return nil, err
		}
		return &v, nil
	default:
		return nil, fmt.Errorf("unknown bankid-se phone status: %s", env.Status)
	}
}

// --- BankID SE endpoint ---

type BankIdSeEndpoint struct {
	client *IdkollenClient
}

func (e *BankIdSeEndpoint) Auth(ctx context.Context, req BankIdSeAuthRequest) (BankIdSeStatus, error) {
	var raw json.RawMessage
	if err := e.client.post(ctx, "/v3/bankid-se/auth", req, &raw); err != nil {
		return nil, err
	}
	return unmarshalBankIdSeStatus(raw)
}

func (e *BankIdSeEndpoint) PhoneAuth(ctx context.Context, req BankIdSePhoneAuthRequest) (BankIdSePhoneStatus, error) {
	var raw json.RawMessage
	if err := e.client.post(ctx, "/v3/bankid-se/phone/auth", req, &raw); err != nil {
		return nil, err
	}
	return unmarshalBankIdSePhoneStatus(raw)
}

func (e *BankIdSeEndpoint) Sign(ctx context.Context, req BankIdSeSignRequest) (BankIdSeStatus, error) {
	var raw json.RawMessage
	if err := e.client.post(ctx, "/v3/bankid-se/sign", req, &raw); err != nil {
		return nil, err
	}
	return unmarshalBankIdSeStatus(raw)
}

func (e *BankIdSeEndpoint) PhoneSign(ctx context.Context, req BankIdSePhoneSignRequest) (BankIdSePhoneStatus, error) {
	var raw json.RawMessage
	if err := e.client.post(ctx, "/v3/bankid-se/phone/sign", req, &raw); err != nil {
		return nil, err
	}
	return unmarshalBankIdSePhoneStatus(raw)
}

func (e *BankIdSeEndpoint) Verify(ctx context.Context, req BankIdSeVerifyRequest) (*BankIdSeVerifyResponse, error) {
	var out BankIdSeVerifyResponse
	if err := e.client.post(ctx, "/v3/bankid-se/verify", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (e *BankIdSeEndpoint) AgeVerification(ctx context.Context, req AgeVerificationRequest) (AgeVerificationStatus, error) {
	var raw json.RawMessage
	if err := e.client.post(ctx, "/v3/bankid-se/age-verification", req, &raw); err != nil {
		return nil, err
	}
	return unmarshalAgeVerificationStatus(raw)
}

func (e *BankIdSeEndpoint) AuthStatus(ctx context.Context, id string) (BankIdSeStatus, error) {
	var raw json.RawMessage
	if err := e.client.get(ctx, "/v3/bankid-se/auth/"+id, &raw); err != nil {
		return nil, err
	}
	return unmarshalBankIdSeStatus(raw)
}

func (e *BankIdSeEndpoint) SignStatus(ctx context.Context, id string) (BankIdSeStatus, error) {
	var raw json.RawMessage
	if err := e.client.get(ctx, "/v3/bankid-se/sign/"+id, &raw); err != nil {
		return nil, err
	}
	return unmarshalBankIdSeStatus(raw)
}

func (e *BankIdSeEndpoint) AgeVerificationStatus(ctx context.Context, id string) (AgeVerificationStatus, error) {
	var raw json.RawMessage
	if err := e.client.get(ctx, "/v3/bankid-se/age-verification/"+id, &raw); err != nil {
		return nil, err
	}
	return unmarshalAgeVerificationStatus(raw)
}

func (e *BankIdSeEndpoint) CancelAuth(ctx context.Context, id string) error {
	return e.client.delete(ctx, "/v3/bankid-se/auth/"+id)
}

func (e *BankIdSeEndpoint) CancelSign(ctx context.Context, id string) error {
	return e.client.delete(ctx, "/v3/bankid-se/sign/"+id)
}

func (e *BankIdSeEndpoint) CancelAgeVerification(ctx context.Context, id string) error {
	return e.client.delete(ctx, "/v3/bankid-se/age-verification/"+id)
}

func (e *BankIdSeEndpoint) WaitForAuth(ctx context.Context, id string, opts PollOptions) (BankIdSeStatus, error) {
	return pollBankIdSe(ctx, opts, func() (BankIdSeStatus, error) { return e.AuthStatus(ctx, id) })
}

func (e *BankIdSeEndpoint) WaitForSign(ctx context.Context, id string, opts PollOptions) (BankIdSeStatus, error) {
	return pollBankIdSe(ctx, opts, func() (BankIdSeStatus, error) { return e.SignStatus(ctx, id) })
}

func (e *BankIdSeEndpoint) WaitForAgeVerification(ctx context.Context, id string, opts PollOptions) (AgeVerificationStatus, error) {
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

func pollBankIdSe(ctx context.Context, opts PollOptions, fn func() (BankIdSeStatus, error)) (BankIdSeStatus, error) {
	interval := opts.Interval
	if interval == 0 {
		interval = 2 * time.Second
	}
	for {
		status, err := fn()
		if err != nil {
			return nil, &WaitError{Err: err}
		}
		if _, ok := status.(*BankIdSePending); !ok {
			return status, nil
		}
		select {
		case <-ctx.Done():
			return nil, &WaitError{Timeout: true}
		case <-time.After(interval):
		}
	}
}

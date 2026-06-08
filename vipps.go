package idkollen

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// --- Vipps request types ---

type VippsAuthRequest struct {
	RedirectUrl    *string `json:"redirectUrl,omitempty"`
	RequestSsn     *bool   `json:"requestSsn,omitempty"`
	RequestPhone   *bool   `json:"requestPhone,omitempty"`
	RequestEmail   *bool   `json:"requestEmail,omitempty"`
	RequestAddress *bool   `json:"requestAddress,omitempty"`
	RefId          *string `json:"refId,omitempty"`
	AppCallbackUri *string `json:"appCallbackUri,omitempty"`
}

type VippsBackchannelAuthRequest struct {
	Phone          string  `json:"phone"`
	RequestSsn     *bool   `json:"requestSsn,omitempty"`
	RequestEmail   *bool   `json:"requestEmail,omitempty"`
	RequestAddress *bool   `json:"requestAddress,omitempty"`
	CallbackUrl    *string `json:"callbackUrl,omitempty"`
	RefId          *string `json:"refId,omitempty"`
}

// --- Vipps status types ---

type VippsStatus interface {
	isVippsStatus()
}

type VippsPending struct {
	Id    string  `json:"id"`
	RefId *string `json:"refId"`
	Url   *string `json:"url"`
}

type VippsCompleted struct {
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

type VippsFailed struct {
	Id    string  `json:"id"`
	RefId *string `json:"refId"`
	Error string  `json:"error"`
}

func (*VippsPending) isVippsStatus()   {}
func (*VippsCompleted) isVippsStatus() {}
func (*VippsFailed) isVippsStatus()    {}

type vippsStatusEnvelope struct {
	Status string `json:"status"`
}

func unmarshalVippsStatus(data []byte) (VippsStatus, error) {
	var env vippsStatusEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, err
	}
	switch env.Status {
	case "PENDING":
		var v VippsPending
		if err := json.Unmarshal(data, &v); err != nil {
			return nil, err
		}
		return &v, nil
	case "COMPLETED":
		var v VippsCompleted
		if err := json.Unmarshal(data, &v); err != nil {
			return nil, err
		}
		return &v, nil
	case "FAILED":
		var v VippsFailed
		if err := json.Unmarshal(data, &v); err != nil {
			return nil, err
		}
		return &v, nil
	default:
		return nil, fmt.Errorf("unknown vipps status: %s", env.Status)
	}
}

// --- Vipps endpoint ---

type VippsEndpoint struct {
	client *IdkollenClient
}

func (e *VippsEndpoint) Auth(ctx context.Context, req VippsAuthRequest) (VippsStatus, error) {
	var raw json.RawMessage
	if err := e.client.post(ctx, "/v3/vipps/auth", req, &raw); err != nil {
		return nil, err
	}
	return unmarshalVippsStatus(raw)
}

func (e *VippsEndpoint) BackchannelAuth(ctx context.Context, req VippsBackchannelAuthRequest) (VippsStatus, error) {
	var raw json.RawMessage
	if err := e.client.post(ctx, "/v3/vipps/backchannel/auth", req, &raw); err != nil {
		return nil, err
	}
	return unmarshalVippsStatus(raw)
}

func (e *VippsEndpoint) AuthStatus(ctx context.Context, id string) (VippsStatus, error) {
	var raw json.RawMessage
	if err := e.client.get(ctx, "/v3/vipps/auth/"+id, &raw); err != nil {
		return nil, err
	}
	return unmarshalVippsStatus(raw)
}

func (e *VippsEndpoint) CancelAuth(ctx context.Context, id string) error {
	return e.client.delete(ctx, "/v3/vipps/auth/"+id)
}

func (e *VippsEndpoint) WaitForAuth(ctx context.Context, id string, opts PollOptions) (VippsStatus, error) {
	interval := opts.Interval
	if interval == 0 {
		interval = 2 * time.Second
	}
	for {
		status, err := e.AuthStatus(ctx, id)
		if err != nil {
			return nil, &WaitError{Err: err}
		}
		if _, ok := status.(*VippsPending); !ok {
			return status, nil
		}
		select {
		case <-ctx.Done():
			return nil, &WaitError{Timeout: true}
		case <-time.After(interval):
		}
	}
}

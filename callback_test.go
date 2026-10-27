package identisure_test

import (
	"testing"

	identisure "github.com/identisure/idk-clients/go/identisure"
)

const (
	callbackBody      = `{"id":"abc","status":"COMPLETED"}`
	callbackSecret    = "test-secret"
	callbackSignature = "Ag2m3etnW31OP585jvcrvOupK7Vgsrzcq17KxKvkIOs="
)

func TestVerifyCallbackSignature_Valid(t *testing.T) {
	if !identisure.VerifyCallbackSignature([]byte(callbackBody), callbackSignature, callbackSecret) {
		t.Error("expected valid signature")
	}
}

func TestVerifyCallbackSignature_TamperedBody(t *testing.T) {
	if identisure.VerifyCallbackSignature([]byte(`{"id":"abd","status":"COMPLETED"}`), callbackSignature, callbackSecret) {
		t.Error("expected tampered body to be rejected")
	}
}

func TestVerifyCallbackSignature_Malformed(t *testing.T) {
	if identisure.VerifyCallbackSignature([]byte(callbackBody), "not base64!", callbackSecret) {
		t.Error("expected malformed signature to be rejected")
	}
}

package cmd

import (
	"errors"
	"testing"

	"github.com/xiaoxiaolin0918/nacos-cli/internal/output"
)

func TestClassifyErrorCodes(t *testing.T) {
	b := classifyError(errors.New("NACOS_SERVER not set; copy .env"), "")
	if b.Code != output.CodeServerUnset {
		t.Fatalf("%+v", b)
	}
	b = classifyError(errors.New("HTTP 401: unauthorized"), "")
	if b.Type != "http" || b.Code != "http_401" {
		t.Fatalf("%+v", b)
	}
	b = classifyError(errors.New("login failed HTTP 403: no"), "")
	if b.Code != "http_403" {
		t.Fatalf("%+v", b)
	}
	b = classifyError(errors.New("refusing high-risk-write without --yes"), "high-risk-write")
	if b.Code != output.CodeConfirmationRequired {
		t.Fatalf("%+v", b)
	}
}

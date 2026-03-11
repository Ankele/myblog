package handler

import (
	"net/http"
	"testing"
)

func TestEnforceStateChangingRequest(t *testing.T) {
	req, err := http.NewRequest(http.MethodPost, "/api/admin/login", nil)
	if err != nil {
		t.Fatalf("unexpected request build error: %v", err)
	}

	if err := enforceStateChangingRequest(req); err == nil {
		t.Fatal("expected missing csrf header to be rejected")
	}

	req.Header.Set(csrfHeaderName, csrfHeaderValue)
	if err := enforceStateChangingRequest(req); err != nil {
		t.Fatalf("expected csrf-like header to pass, got %v", err)
	}
}

func TestClientIP(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatalf("unexpected request build error: %v", err)
	}
	req.RemoteAddr = "127.0.0.1:8080"

	if got := clientIP(req); got != "127.0.0.1" {
		t.Fatalf("unexpected client ip: %s", got)
	}
}

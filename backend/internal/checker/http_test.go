package checker

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPChecker_Healthy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := NewHTTPChecker(srv.URL, false)
	status, code, elapsed, msg := c.Check(context.Background())

	if status != StatusHealthy {
		t.Fatalf("expected Healthy, got %s", status)
	}
	if code != 200 {
		t.Fatalf("expected 200, got %d", code)
	}
	if elapsed < 0 {
		t.Fatal("expected non-negative elapsed time")
	}
	if msg != "" {
		t.Fatalf("expected empty message, got %q", msg)
	}
}

func TestHTTPChecker_Unhealthy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := NewHTTPChecker(srv.URL, false)
	status, code, _, _ := c.Check(context.Background())

	if status != StatusUnhealthy {
		t.Fatalf("expected Unhealthy, got %s", status)
	}
	if code != 500 {
		t.Fatalf("expected 500, got %d", code)
	}
}

func TestHTTPChecker_TLSDefaultRejectsSelfSigned(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := NewHTTPChecker(srv.URL, false)
	status, _, _, msg := c.Check(context.Background())

	if status != StatusUnhealthy {
		t.Fatalf("expected Unhealthy for self-signed cert with default verify, got %s", status)
	}
	// The error should mention certificate verification.
	if !strings.Contains(msg, "certificate") && !strings.Contains(msg, "x509") {
		t.Fatalf("expected TLS error in message, got %q", msg)
	}
}

func TestHTTPChecker_SkipVerifyAcceptsSelfSigned(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := NewHTTPChecker(srv.URL, true)
	status, code, _, msg := c.Check(context.Background())

	if status != StatusHealthy {
		t.Fatalf("expected Healthy with skip_verify, got %s (msg: %s)", status, msg)
	}
	if code != 200 {
		t.Fatalf("expected 200, got %d", code)
	}
}

func TestHTTPChecker_SkipVerifyTransportConfig(t *testing.T) {
	// Verify that skip_verify=false does NOT set InsecureSkipVerify.
	cDefault := NewHTTPChecker("https://example.com", false)
	if cDefault.client.Transport != nil {
		t.Fatal("expected nil transport for default checker")
	}

	// Verify that skip_verify=true sets InsecureSkipVerify.
	cSkip := NewHTTPChecker("https://example.com", true)
	transport, ok := cSkip.client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("expected *http.Transport when skip_verify is true")
	}
	if !transport.TLSClientConfig.InsecureSkipVerify {
		t.Fatal("expected InsecureSkipVerify to be true")
	}

}

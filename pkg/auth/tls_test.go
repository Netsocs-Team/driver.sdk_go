package auth

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Netsocs-Team/driver.sdk_go/internal/testpki"
	"github.com/Netsocs-Team/driver.sdk_go/pkg/httpx"
)

func resetTLS(t *testing.T) {
	t.Cleanup(func() {
		Activate(nil)
		_ = httpx.SecureCentral("", nil)
	})
}

// Discovery, token and DriverHub calls all go over TLS verified against the
// CA in ca_file (relative to driver.netsocs.json).
func TestSetupPinnedCA(t *testing.T) {
	resetTLS(t)
	dir := t.TempDir()
	ca := testpki.NewCA(t, "central root")
	c := newCentralTLS(t, &writeKey(t, dir).PublicKey, ca.Server(t, "127.0.0.1"))
	if err := os.WriteFile(filepath.Join(dir, "central-ca.pem"), ca.PEM, 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := c.cfg()
	cfg.CAFile, cfg.CAMode = "central-ca.pem", "pinned"

	mode, _, err := Setup(context.Background(), cfg, dir, driverKey, c.hub(), t.Logf)
	if err != nil || mode != ModeJWT {
		t.Fatalf("mode %s err %v", mode, err)
	}
	if h := c.callHub(t); h.Get("Authorization") != "Bearer jwt-1" {
		t.Fatalf("DriverHub got Authorization=%q", h.Get("Authorization"))
	}
}

// A central whose certificate is not from the pinned CA is never trusted:
// discovery keeps failing (and retrying) instead of picking a mode.
func TestSetupPinnedCARejectsAnotherCA(t *testing.T) {
	resetTLS(t)
	dir := t.TempDir()
	c := newCentralTLS(t, &writeKey(t, dir).PublicKey, testpki.NewCA(t, "attacker").Server(t, "127.0.0.1"))
	_ = os.WriteFile(filepath.Join(dir, "central-ca.pem"), testpki.NewCA(t, "central root").PEM, 0o644)
	cfg := c.cfg()
	cfg.CAFile, cfg.CAMode = "central-ca.pem", "pinned"

	var logs strings.Builder
	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()
	_, _, err := Setup(ctx, cfg, dir, driverKey, c.hub(), func(f string, a ...any) {
		logs.WriteString(f)
		for _, v := range a {
			if e, ok := v.(error); ok {
				logs.WriteString(e.Error())
			}
		}
	})
	if !errors.Is(err, context.DeadlineExceeded) || !strings.Contains(logs.String(), "x509") {
		t.Fatalf("err %v, logs %q", err, logs.String())
	}
	if atomic.LoadInt32(&c.tokens) != 0 {
		t.Fatal("a token was requested from an untrusted central")
	}
}

// Settings that cannot be honoured stop the driver before any request,
// never falling back to unverified TLS.
func TestSetupRejectsUnusableTLSSettings(t *testing.T) {
	dir := t.TempDir()
	writeKey(t, dir)
	_ = os.WriteFile(filepath.Join(dir, "bad.pem"), []byte("not a certificate"), 0o644)
	hub := "https://central.example/api/netsocs/dh"
	tok := "https://central.example/auth/realms/netsocs/protocol/openid-connect/token"
	cases := map[string]struct {
		hub, tok, file, mode string
	}{
		"pinned without ca_file": {hub, tok, "", "pinned"},
		"missing ca_file":        {hub, tok, "nope.pem", "pinned"},
		"ca_file not PEM":        {hub, tok, "bad.pem", "pinned"},
		"unknown ca_mode":        {hub, tok, "", "trust-me"},
		"http driver_hub_host":   {"http://central.example/api/netsocs/dh", tok, "", "public"},
		"http token_url":         {hub, "http://central.example/auth/realms/netsocs/protocol/openid-connect/token", "", "public"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			resetTLS(t)
			cfg := &Config{ClientID: "drv-" + strings.Repeat("a", 32), TokenURL: tc.tok, KeyFile: "driver.key",
				CAFile: tc.file, CAMode: tc.mode}
			_, _, err := Setup(context.Background(), cfg, dir, driverKey, tc.hub, t.Logf)
			if !errors.Is(err, ErrTLSConfig) {
				t.Fatalf("got %v, want ErrTLSConfig", err)
			}
		})
	}
}

// No ca_mode: exactly v0.8.0-rc.1, a central with a private CA still works.
func TestSetupWithoutCAModeKeepsLegacyTLS(t *testing.T) {
	resetTLS(t)
	dir := t.TempDir()
	c := newCentralTLS(t, &writeKey(t, dir).PublicKey, testpki.NewCA(t, "private root").Server(t, "127.0.0.1"))
	cfg := c.cfg()
	cfg.CAFile = "ignored.pem"
	if mode, _, err := Setup(context.Background(), cfg, dir, driverKey, c.hub(), t.Logf); err != nil || mode != ModeJWT {
		t.Fatalf("mode %s err %v", mode, err)
	}
}

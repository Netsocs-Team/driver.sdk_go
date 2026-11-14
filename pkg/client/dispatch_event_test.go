package client

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/Netsocs-Team/driver.sdk_go/pkg/objects"
)

var uuidV7 = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func dispatchAgainst(t *testing.T, status int, ev objects.Event) (objects.NewEventRequestBodySchema, error) {
	t.Helper()
	var got objects.NewEventRequestBodySchema
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&got)
		w.WriteHeader(status)
		_, _ = w.Write([]byte(`{"id":"x"}`))
	}))
	defer srv.Close()
	c := &NetsocsDriverClient{token: "t", driverHubHost: srv.URL}
	_, err := c.DispatchEvent("alarm", "fire", ev)
	return got, err
}

func TestDispatchEventStoredSendsIDAndTime(t *testing.T) {
	got, err := dispatchAgainst(t, http.StatusCreated, objects.Event{ObjectIDs: []string{"o1"}})
	if err != nil {
		t.Fatalf("201 must be success: %v", err)
	}
	if !uuidV7.MatchString(got.ID) {
		t.Fatalf("id %q is not a UUIDv7", got.ID)
	}
	if _, err := time.Parse(time.RFC3339Nano, got.OccurredAt); err != nil {
		t.Fatalf("occurred_at %q: %v", got.OccurredAt, err)
	}
	if got.EventType != "alarm.fire" || len(got.Rels) != 1 || got.Rels[0] != "/objects/o1" {
		t.Fatalf("body: %+v", got)
	}
}

func TestDispatchEventKeepsTheCallersIDAndTime(t *testing.T) {
	at := time.Date(2026, 11, 13, 20, 0, 0, 0, time.UTC)
	got, err := dispatchAgainst(t, http.StatusOK, objects.Event{ID: "0192f0a4-7c1e-7b3a-9d2e-4f5a6b7c8d9e", OccurredAt: at})
	if err != nil {
		t.Fatalf("200 (a retry) must be success: %v", err)
	}
	if got.ID != "0192f0a4-7c1e-7b3a-9d2e-4f5a6b7c8d9e" || got.OccurredAt != "2026-11-13T20:00:00Z" {
		t.Fatalf("body: %+v", got)
	}
}

func TestDispatchEventNotStoredIsAnError(t *testing.T) {
	for _, tc := range []struct {
		status    int
		retryable bool
	}{{503, true}, {429, true}, {409, false}, {400, false}} {
		got, err := dispatchAgainst(t, tc.status, objects.Event{})
		var nse *objects.EventNotStoredError
		if !errors.As(err, &nse) {
			t.Fatalf("%d: want EventNotStoredError, got %v", tc.status, err)
		}
		if nse.StatusCode != tc.status || nse.Retryable() != tc.retryable || nse.ID != got.ID {
			t.Fatalf("%d: %+v (retryable %v)", tc.status, nse, nse.Retryable())
		}
	}
}

func TestDispatchEventNoAnswerIsRetryable(t *testing.T) {
	c := &NetsocsDriverClient{token: "t", driverHubHost: "http://127.0.0.1:1"}
	_, err := c.DispatchEvent("alarm", "fire", objects.Event{})
	var nse *objects.EventNotStoredError
	if !errors.As(err, &nse) || nse.StatusCode != 0 || !nse.Retryable() {
		t.Fatalf("want a retryable EventNotStoredError, got %v", err)
	}
}

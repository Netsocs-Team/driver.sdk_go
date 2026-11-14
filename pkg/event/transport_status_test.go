package event

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPostEventReportsTheServerAnswer(t *testing.T) {
	status := http.StatusOK
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write([]byte("store down"))
	}))
	defer srv.Close()

	if err := postEvent(srv.URL, 0, "k", strings.NewReader("{}")); err != nil {
		t.Fatalf("200: %v", err)
	}
	status = http.StatusServiceUnavailable
	err := postEvent(srv.URL, 0, "k", strings.NewReader("{}"))
	if err == nil || !strings.Contains(err.Error(), "503") || !strings.Contains(err.Error(), "store down") {
		t.Fatalf("503 must be an error with the answer, got %v", err)
	}
}

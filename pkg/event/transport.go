package event

import (
	"fmt"
	"io"
	"strings"

	"github.com/Netsocs-Team/driver.sdk_go/pkg/httpx"
	"github.com/Netsocs-Team/driver.sdk_go/pkg/objects"
)

// eventURL builds the misc-event endpoint for host/port.
//
// host may already carry a scheme (e.g. "https://nutresa.netsocs.com"), in which
// case it is used as-is and the port is ignored; otherwise "http://host:port" is
// assembled as it always was.
func eventURL(host string, port int, eventKey string) string {
	host = strings.TrimSuffix(strings.TrimSpace(host), "/")
	if strings.HasPrefix(host, "http://") || strings.HasPrefix(host, "https://") {
		return fmt.Sprintf("%s/v1/topologia/misc/%s", host, eventKey)
	}
	return fmt.Sprintf("http://%s:%d/v1/topologia/misc/%s", host, port, eventKey)
}

// eventClient bounds each send (objects.EventSendTimeout) and shares the SDK
// transport, so HTTPS hosts honour the TLS configuration in pkg/httpx.
var eventClient = httpx.NewClient(objects.EventSendTimeout)

// postEvent sends the body and reports whether the server accepted it: a
// non-2xx answer is an error, as is no answer within EventSendTimeout.
func postEvent(host string, port int, eventKey string, body io.Reader) error {
	res, err := eventClient.Post(eventURL(host, port, eventKey), "application/json", body)
	if err != nil {
		return err
	}
	// Drain and close so the connection returns to the pool.
	defer res.Body.Close()
	msg, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
	_, _ = io.Copy(io.Discard, res.Body)
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return fmt.Errorf("event %s not accepted: server answered %d: %s", eventKey, res.StatusCode, strings.TrimSpace(string(msg)))
	}
	return nil
}

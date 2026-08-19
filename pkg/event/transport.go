package event

import (
	"fmt"
	"io"
	"strings"

	"github.com/Netsocs-Team/driver.sdk_go/pkg/httpx"
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

// postEvent sends the body through the shared SDK client, so HTTPS hosts honour
// the TLS configuration in pkg/httpx.
func postEvent(host string, port int, eventKey string, body io.Reader) error {
	res, err := httpx.Client().Post(eventURL(host, port, eventKey), "application/json", body)
	if err != nil {
		return err
	}
	// Drain and close so the connection returns to the pool.
	defer res.Body.Close()
	_, _ = io.Copy(io.Discard, res.Body)
	return nil
}

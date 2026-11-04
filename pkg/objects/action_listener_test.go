package objects

import (
	"github.com/Netsocs-Team/driver.sdk_go/pkg/httpx"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/Netsocs-Team/driver.sdk_go/internal/eventbus"
	"github.com/goccy/go-json"
	"github.com/gorilla/websocket"
)

// fakeObjectsHub accepts objects/ws connections and hands each one to the test.
func fakeObjectsHub(t *testing.T) (*httptest.Server, <-chan *websocket.Conn) {
	t.Helper()
	conns := make(chan *websocket.Conn, 4)
	upgrader := websocket.Upgrader{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		conns <- c
	}))
	t.Cleanup(srv.Close)
	return srv, conns
}

func nextConn(t *testing.T, conns <-chan *websocket.Conn) *websocket.Conn {
	t.Helper()
	select {
	case c := <-conns:
		return c
	case <-time.After(10 * time.Second):
		t.Fatal("driver did not (re)connect to objects/ws")
		return nil
	}
}

func expectSubscription(t *testing.T, c *websocket.Conn, domain string) {
	t.Helper()
	_ = c.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, raw, err := c.ReadMessage()
	if err != nil {
		t.Fatalf("reading subscription: %v", err)
	}
	var msg wsMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		t.Fatalf("decoding subscription: %v", err)
	}
	if msg.EventType != "REQUEST_SUBSCRIPTION_TO_DOMAIN" || msg.Domain != domain {
		t.Fatalf("got %+v, want subscription to %s", msg, domain)
	}
}

func TestListenActionRequestsReconnectsAndResubscribes(t *testing.T) {
	srv, conns := fakeObjectsHub(t)
	domain := "test_reconnect.video_channel"

	ctrl := &objectController{driverhub_host: srv.URL, driver_key: "key", actions: newActionsChannel()}
	// Registered before the socket exists: must not be lost.
	eventbus.Pubsub.Publish("SUBSCRIBE_OBJECTS_COMMANDS_LISTENING", struct{ Domain string }{Domain: domain})

	received := make(chan string, 4)
	handler := func(data interface{}) {
		if m, ok := data.(map[string]interface{}); ok && m["domain"] == domain {
			received <- m["id"].(string)
		}
	}
	eventbus.Pubsub.Subscribe("REQUEST_ACTION_EXECUTION", handler)
	t.Cleanup(func() { eventbus.Pubsub.Unsubscribe("REQUEST_ACTION_EXECUTION", handler) })

	go ctrl.ListenActionRequests()

	first := nextConn(t, conns)
	expectSubscription(t, first, domain)
	first.Close() // simulate the network dropping the socket

	second := nextConn(t, conns)
	expectSubscription(t, second, domain)

	action := `{"event_type":"REQUEST_ACTION_EXECUTION","data":{"id":"exec-1","domain":"` + domain + `","action":"video_channel.action.snapshot"}}`
	if err := second.WriteMessage(websocket.TextMessage, []byte(action)); err != nil {
		t.Fatalf("sending action: %v", err)
	}
	select {
	case id := <-received:
		if id != "exec-1" {
			t.Fatalf("got execution %s", id)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("action sent after reconnect never reached the driver")
	}
}

func TestDomainAddedWhileConnectedIsSubscribedImmediately(t *testing.T) {
	srv, conns := fakeObjectsHub(t)
	domain := "test_live_subscribe.switch"

	ctrl := &objectController{driverhub_host: srv.URL, driver_key: "key", actions: newActionsChannel()}
	go ctrl.ListenActionRequests()
	c := nextConn(t, conns)

	// Wait until the socket is attached before publishing the domain.
	deadline := time.Now().Add(5 * time.Second)
	for {
		ctrl.actions.mu.Lock()
		attached := ctrl.actions.conn != nil
		ctrl.actions.mu.Unlock()
		if attached {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("socket never attached")
		}
		time.Sleep(10 * time.Millisecond)
	}

	eventbus.Pubsub.Publish("SUBSCRIBE_OBJECTS_COMMANDS_LISTENING", struct{ Domain string }{Domain: domain})
	// Other tests share the global bus; skip unrelated domains they replay.
	for i := 0; i < 10; i++ {
		_ = c.SetReadDeadline(time.Now().Add(5 * time.Second))
		_, raw, err := c.ReadMessage()
		if err != nil {
			t.Fatalf("reading subscription: %v", err)
		}
		if strings.Contains(string(raw), domain) {
			return
		}
	}
	t.Fatalf("no subscription for %s", domain)
}

// With a driver identity the objects/ws handshake carries its token, like
// every other call to the DriverHub.
func TestActionsSocketCarriesTheDriverIdentity(t *testing.T) {
	got := make(chan string, 1)
	upgrader := websocket.Upgrader{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case got <- r.Header.Get("Authorization"):
		default:
		}
		if c, err := upgrader.Upgrade(w, r, nil); err == nil {
			defer c.Close()
			_, _, _ = c.ReadMessage()
		}
	}))
	t.Cleanup(srv.Close)
	httpx.SetAuthHook(func(_ *url.URL, h http.Header) { h.Set("Authorization", "Bearer drv-token") })
	t.Cleanup(func() { httpx.SetAuthHook(nil) })

	ctrl := &objectController{driverhub_host: srv.URL, driver_key: "key", actions: newActionsChannel()}
	go ctrl.ListenActionRequests()

	select {
	case h := <-got:
		if h != "Bearer drv-token" {
			t.Fatalf("Authorization = %q, want the driver token", h)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("driver never opened objects/ws")
	}
}

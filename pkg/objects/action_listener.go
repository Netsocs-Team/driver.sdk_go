package objects

import (
	"net/http"
	"sync"
	"time"

	"github.com/Netsocs-Team/driver.sdk_go/internal/eventbus"
	"github.com/Netsocs-Team/driver.sdk_go/pkg/httpx"
	"github.com/Netsocs-Team/driver.sdk_go/pkg/logger"
	"github.com/Netsocs-Team/driver.sdk_go/pkg/tools"
	"github.com/goccy/go-json"
	"github.com/gorilla/websocket"
)

// The objects/ws channel is the only path DriverHub uses to deliver action
// requests (snapshots, PTZ, video clips...) to the driver. It used to be
// dialled once: when the socket dropped, the listener returned silently and the
// driver stayed "online" (config channel alive, events flowing) while every
// action was lost until the process restarted. The listener now keeps the
// channel alive for the life of the process.
const (
	// actionsWSPongWait is how long the socket may stay silent before it is
	// considered dead. DriverHub pings every 54s and answers our pings, so a
	// healthy socket always sees traffic well inside this window.
	actionsWSPongWait = 90 * time.Second
	// actionsWSPingPeriod must stay below actionsWSPongWait.
	actionsWSPingPeriod = 30 * time.Second
	actionsWSWriteWait  = 10 * time.Second
	actionsWSMinBackoff = 1 * time.Second
	actionsWSMaxBackoff = 30 * time.Second
)

// actionsChannel tracks the domains the driver must be subscribed to and the
// live socket they are sent on. DriverHub keeps subscriptions per connection,
// so every reconnect has to replay them.
type actionsChannel struct {
	mu      sync.Mutex
	domains []string
	conn    *websocket.Conn
}

func newActionsChannel() *actionsChannel {
	ch := &actionsChannel{}
	// Registered at construction time, before any object exists, so a domain
	// published while the socket is still dialling is remembered instead of
	// being dropped.
	eventbus.Pubsub.Subscribe("SUBSCRIBE_OBJECTS_COMMANDS_LISTENING", func(data interface{}) {
		if domain := domainFromSubscribeEvent(data); domain != "" {
			ch.addDomain(domain)
		}
	})
	return ch
}

func domainFromSubscribeEvent(data interface{}) string {
	if d, ok := data.(struct{ Domain string }); ok {
		return d.Domain
	}
	raw, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	var v struct {
		Domain string `json:"Domain"`
	}
	if json.Unmarshal(raw, &v) != nil {
		return ""
	}
	return v.Domain
}

func (ch *actionsChannel) addDomain(domain string) {
	ch.mu.Lock()
	defer ch.mu.Unlock()
	for _, d := range ch.domains {
		if d == domain {
			return
		}
	}
	ch.domains = append(ch.domains, domain)
	if ch.conn != nil {
		if err := writeDomainSubscription(ch.conn, domain); err != nil {
			logger.Logger().Errorf("actions websocket: subscribe to domain %s failed: %v", domain, err)
		}
	}
}

// attach makes conn the active socket and replays every known subscription.
// Writes happen under ch.mu, which is the single writer lock for data frames.
func (ch *actionsChannel) attach(conn *websocket.Conn) error {
	ch.mu.Lock()
	defer ch.mu.Unlock()
	for _, d := range ch.domains {
		if err := writeDomainSubscription(conn, d); err != nil {
			return err
		}
	}
	ch.conn = conn
	return nil
}

func (ch *actionsChannel) detach(conn *websocket.Conn) {
	ch.mu.Lock()
	defer ch.mu.Unlock()
	if ch.conn == conn {
		ch.conn = nil
	}
}

func writeDomainSubscription(conn *websocket.Conn, domain string) error {
	_ = conn.SetWriteDeadline(time.Now().Add(actionsWSWriteWait))
	return conn.WriteJSON(wsMessage{EventType: "REQUEST_SUBSCRIPTION_TO_DOMAIN", Domain: domain})
}

// ListenActionRequests implements ObjectController. It never returns: the
// actions socket is redialled with exponential backoff whenever it drops or
// goes silent.
func (o *objectController) ListenActionRequests() error {
	if o.actions == nil {
		o.actions = newActionsChannel()
	}
	backoff := actionsWSMinBackoff
	for {
		connected, err := o.runActionsConnection()
		if connected {
			backoff = actionsWSMinBackoff
		}
		logger.Logger().Warnf("actions websocket disconnected (%v); reconnecting in %s", err, backoff)
		time.Sleep(backoff)
		if backoff *= 2; backoff > actionsWSMaxBackoff {
			backoff = actionsWSMaxBackoff
		}
	}
}

// runActionsConnection serves one objects/ws session until it fails. connected
// reports whether the socket was established, to reset the backoff.
func (o *objectController) runActionsConnection() (connected bool, err error) {
	url := tools.ConvertToWebSocketURL(o.driverhub_host, "objects/ws")
	// DialWebsocket also adds the driver's identity token when it has one.
	conn, _, err := httpx.DialWebsocket(url, http.Header{
		"X-Auth-Token": []string{o.token},
	})
	if err != nil {
		return false, err
	}
	defer conn.Close()

	extend := func() { _ = conn.SetReadDeadline(time.Now().Add(actionsWSPongWait)) }
	extend()
	conn.SetPongHandler(func(string) error { extend(); return nil })
	conn.SetPingHandler(func(appData string) error {
		extend()
		err := conn.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(actionsWSWriteWait))
		if err == websocket.ErrCloseSent {
			return nil
		}
		return err
	})

	if err := o.actions.attach(conn); err != nil {
		return true, err
	}
	defer o.actions.detach(conn)
	logger.Logger().Info("actions websocket connected")

	done := make(chan struct{})
	defer close(done)
	go func() {
		ticker := time.NewTicker(actionsWSPingPeriod)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				// WriteControl is safe to call concurrently with WriteJSON.
				if err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(actionsWSWriteWait)); err != nil {
					_ = conn.Close() // unblocks ReadMessage
					return
				}
			}
		}
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			return true, err
		}
		extend()
		msg := wsMessage{}
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}
		if msg.EventType == "REQUEST_ACTION_EXECUTION" {
			go eventbus.Pubsub.Publish("REQUEST_ACTION_EXECUTION", msg.Data)
		}
	}
}

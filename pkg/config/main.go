package config

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"sync"

	"github.com/Netsocs-Team/driver.sdk_go/pkg/tools"
	"github.com/gorilla/websocket"
)

type ConfigMessagePort interface {
	GetDeviceData() (interface{}, error)
}

type ConfigMessage struct {
	DeviceData *ConfigMessageDeviceData `json:"deviceData"`
	ConfigKey  NetsocsConfigKey         `json:"configKey"`
	Value      string                   `json:"value"`
	RequestID  string                   `json:"requestId"`
	rawMessage []byte
}

func (c *ConfigMessage) GetDeviceData() (interface{}, error) {
	return nil, nil
}

func (c *ConfigMessage) GetConfigKey() NetsocsConfigKey {
	return c.ConfigKey
}

func (c *ConfigMessage) GetRawMessage() []byte {
	return c.rawMessage
}

type ConfigMessageDeviceData struct {
	Username         string                 `json:"username"`
	Password         string                 `json:"password"`
	IP               string                 `json:"ip_address_public"`
	Name             string                 `json:"device_name"`
	Port             int                    `json:"port"`
	IsSSL            bool                   `json:"is_ssl"`
	SSLPort          int                    `json:"ssl_port"`
	ID               int                    `json:"id_device"`
	ChildID          string                 `json:"child_id"`
	Extrafields      map[string]interface{} `json:"extrafields"`
	SetVideoEngineID func(string)           `json:"-"`
}

type VideoEngineAdditionalProperties struct {
	APIPort      string `json:"api_port"`
	HLSPort      string `json:"hls_port"`
	Hostname     string `json:"hostname"`
	Offline      string `json:"offline"`
	PlaybackPort string `json:"playback_port"`
	RTSPPort     string `json:"rtsp_port"`
	SiteID       string `json:"site_id"`
	State        string `json:"state"`
	WebrtcPort   string `json:"webrtc_port"`
}

// messages is buffered so the websocket read loop never blocks while handlers
// are busy. With an unbuffered channel and a single consumer, the read loop
// stalls on the second queued message; the DriverHub's health-check PINGs then
// sit unread in the socket, the hub counts them as failures and evicts the
// driver (~2.5 min with default hub settings). The connect-time burst is a few
// messages per device, so this comfortably absorbs fleets of hundreds.
var messages = make(chan *ConfigMessage, 1024)
var responses = make(chan *s_response)

type s_response struct {
	RequestId string `json:"requestId"`
	Data      string `json:"data"`
}

type defaultDataResponse struct {
	Error bool   `json:"error"`
	Msg   string `json:"msg"`
}

// defaultConfigWorkers is how many config handlers may run at the same time.
// Handlers do device I/O (ping, channel discovery) that can take tens of
// seconds per device; running them one at a time makes a multi-device fleet
// take sum-of-all-devices to initialize. Override with NETSOCS_CONFIG_WORKERS.
const defaultConfigWorkers = 8

func configWorkerCount() int {
	if v := os.Getenv("NETSOCS_CONFIG_WORKERS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 1 {
			return n
		}
	}
	return defaultConfigWorkers
}

// startConfigWorkers launches the pool that consumes `messages`. Messages for
// the same device always land on the same worker, so per-device order is
// preserved (requestCreateObjects runs before that device's actionListenEvent);
// messages without device data share worker 0. Different devices initialize in
// parallel, so one slow or unreachable device cannot stall the whole fleet nor
// the websocket read loop.
var startWorkersOnce sync.Once

func startConfigWorkers() {
	startWorkersOnce.Do(func() {
		workers := configWorkerCount()
		queues := make([]chan *ConfigMessage, workers)
		for i := range queues {
			queues[i] = make(chan *ConfigMessage, 256)
			go func(q chan *ConfigMessage) {
				for message := range q {
					handleConfigMessage(message)
				}
			}(queues[i])
		}
		go func() {
			for message := range messages {
				idx := 0
				if message.DeviceData != nil {
					idx = message.DeviceData.ID % workers
					if idx < 0 {
						idx = -idx
					}
				}
				queues[idx] <- message
			}
		}()
	})
}

// handleConfigMessage dispatches one config message to its registered handler
// and pushes the reply onto `responses`.
func handleConfigMessage(message *ConfigMessage) {
	handler := handlersMap[message.ConfigKey]
	if handler == nil {
		sendDefaultResponse(message.RequestID, true, fmt.Sprintf("'%s' not found on the driver", message.ConfigKey))
		return
	}
	response, err := handler(message.Value, message.DeviceData)
	if err != nil {
		sendDefaultResponse(message.RequestID, true, err.Error())
		return
	}
	if response == "" || response == "null" {
		sendDefaultResponse(message.RequestID, false, "OK")
		return
	}
	responses <- &s_response{
		RequestId: message.RequestID,
		Data:      response,
	}
}

func sendDefaultResponse(requestID string, isError bool, msg string) {
	jsondata, err := json.Marshal(&defaultDataResponse{Error: isError, Msg: msg})
	if err != nil {
		fmt.Println("Error in handler:", err)
		return
	}
	responses <- &s_response{
		RequestId: requestID,
		Data:      string(jsondata),
	}
}

// This function will start a websocket listener for all 'configuration' requests
// coming from the DriverHub.
// In the SDK, there is a map of handlers that can be registered for each configuration.
// This function, upon receiving a configuration, will look in the map of handlers
// to see if there is a handler for that configuration. If there is no handler, it will return an error.
// More information here https://.../docs
func ListenConfig(host string, driverKey string, siteId string, token string, driverID string, setVideoEngineID func(string, VideoEngineAdditionalProperties), driverVersion string, driverDocumentation string) error {
	startConfigWorkers()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Convert the URL using the utility function with query parameters
	documentationInBase64 := ""
	if driverDocumentation != "" {
		documentationInBase64 = base64.StdEncoding.EncodeToString([]byte(driverDocumentation))
	}
	path := fmt.Sprintf("ws/v1/config_communication?site_id=%s&driver_id=%s&driver_version=%s&driver_documentation=%s", siteId, driverID, driverVersion, documentationInBase64)
	wsURL := tools.ConvertToWebSocketURL(host, path)

	u, err := url.Parse(wsURL)
	if err != nil {
		return err
	}

	log.Printf("connecting to %s", u.String())

	// Create a custom dialer that accepts self-signed certificates
	dialer := websocket.Dialer{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	c, _, err := dialer.Dial(u.String(), http.Header{
		"Authorization": []string{driverKey},
		"X-Auth-Token":  []string{token},
	})
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			configMessage := &ConfigMessage{}
			err = json.Unmarshal(message, configMessage)
			if err != nil {
				log.Println("unmarshal:", err)
			} else {
				// Handle PING message immediately
				if configMessage.ConfigKey == PING {
					log.Printf("recv PING, responding with PONG")
					responses <- &s_response{
						RequestId: configMessage.RequestID,
						Data:      "pong",
					}
					continue
				}

				messages <- configMessage
			}

			if configMessage.ConfigKey == SAVE_VIDEO_ENGINE {
				type msg struct {
					VideoEngine                     string                          `json:"video_engine"`
					VideoEngineAdditionalProperties VideoEngineAdditionalProperties `json:"video_engine_additional_properties"`
				}

				var msgData msg
				err = json.Unmarshal([]byte(configMessage.Value), &msgData)
				if err != nil {
					log.Println("unmarshal msgData:", err)
				} else {
					if setVideoEngineID != nil {
						setVideoEngineID(msgData.VideoEngine, msgData.VideoEngineAdditionalProperties)
					}
				}
			}

			log.Printf("recv: %s", message)
		}
	}()

	for {
		select {
		case response := <-responses:
			jsondata, err := json.Marshal(response)
			if err != nil {
				log.Println("marshal:", err)
			} else {
				err = c.WriteMessage(websocket.TextMessage, jsondata)
				if err != nil {
					log.Println("write:", err)
					return err
				}
			}

		case <-done:
			return nil
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return err
			}
			<-done

			return nil
		}
	}

}

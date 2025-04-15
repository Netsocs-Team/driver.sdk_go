package config

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"

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
	Username    string                 `json:"username"`
	Password    string                 `json:"password"`
	IP          string                 `json:"ip_address_public"`
	Port        int                    `json:"port"`
	IsSSL       bool                   `json:"is_ssl"`
	SSLPort     int                    `json:"ssl_port"`
	ID          int                    `json:"id_device"`
	ChildID     string                 `json:"child_id"`
	Extrafields map[string]interface{} `json:"extrafields"`
}

var messages = make(chan *ConfigMessage)
var responses = make(chan *s_response)

type s_response struct {
	RequestId string `json:"requestId"`
	Data      string `json:"data"`
}

type defaultDataResponse struct {
	Error bool   `json:"error"`
	Msg   string `json:"msg"`
}

// This function will start a websocket listener for all 'configuration' requests
// coming from the DriverHub.
// In the SDK, there is a map of handlers that can be registered for each configuration.
// This function, upon receiving a configuration, will look in the map of handlers
// to see if there is a handler for that configuration. If there is no handler, it will return an error.
// More information here https://.../docs
func ListenConfig(host string, driverKey string) error {
	go func() {
		for {
			select {
			case message := <-messages:
				handler := handlersMap[message.ConfigKey]
				if handler != nil {
					response, err := handler(message.Value, message.DeviceData)
					if err == nil {
						if response == "" || response == "null" {
							tmp := &defaultDataResponse{
								Error: false,
								Msg:   "OK",
							}
							jsondata, err := json.Marshal(tmp)
							if err != nil {
								fmt.Println("Error in handler:", err)
							} else {
								responses <- &s_response{
									RequestId: message.RequestID,
									Data:      string(jsondata),
								}
							}
						} else {
							responses <- &s_response{
								RequestId: message.RequestID,
								Data:      response,
							}
						}
					} else {
						tmp := &defaultDataResponse{
							Error: true,
							Msg:   err.Error(),
						}
						jsondata, err := json.Marshal(tmp)
						if err != nil {
							fmt.Println("Error in handler:", err)
						} else {
							responses <- &s_response{
								RequestId: message.RequestID,
								Data:      string(jsondata),
							}
						}
					}
				} else {
					tmp := &defaultDataResponse{
						Error: true,
						Msg:   fmt.Sprintf("'%s' not found on the driver", message.ConfigKey),
					}
					jsondata, err := json.Marshal(tmp)
					if err != nil {
						fmt.Println("Error in handler:", err)
					} else {
						responses <- &s_response{
							RequestId: message.RequestID,
							Data:      string(jsondata),
						}
					}
				}
			}
		}
	}()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	if strings.HasPrefix(host, "http") || strings.HasPrefix(host, "https") {
		host = strings.ReplaceAll(host, "https://", "")
		host = strings.ReplaceAll(host, "http://", "")
	}
	u, err := url.Parse(fmt.Sprintf("ws://%s/ws/v1/config_communication", host))
	if err != nil {
		return err
	}
	// u := url.URL{Scheme: "ws", Host: host, Path: "/ws/v1/config_communication"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), http.Header{
		"Authorization": []string{driverKey},
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
				messages <- configMessage
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
			select {
			case <-done:
			}
			return nil
		}
	}

}

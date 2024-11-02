package event

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"strconv"
)

// Deprecated: Use `Dispatcher` instead.
type EventDispatcher struct {
	host string
	port *int
	key  string
}

const DEFAULT_EVENT_PORT = 3070

func NewEventDispatcher(host string, key string, port *int) *EventDispatcher {
	// if host == "xxx/" quite / from the end
	parsed_host := host
	if host[len(host)-1] == '/' {
		parsed_host = host[:len(host)-1]
	}
	return &EventDispatcher{
		host: parsed_host,
		port: port,
		key:  key,
	}
}

func (e *EventDispatcher) SetPort(port *int) {
	e.port = port
}

func (e *EventDispatcher) Dispatch(eventKey string, deviceID int) error {
	bodyString := fmt.Sprintf(`{"deviceId": %d }`, deviceID)
	body := []byte(bodyString)
	bodyReader := bytes.NewReader(body)

	_, err := http.Post(e.CreateUrlEvent(eventKey), "application/json", bodyReader)
	return err
}
func (e *EventDispatcher) DispatchWithOriginalValue(eventKey string, deviceID int, originalValue string) error {
	bodyString := fmt.Sprintf(`{"deviceId": %d, "originalValue": "%s" }`, deviceID, originalValue)
	body := []byte(bodyString)
	bodyReader := bytes.NewReader(body)

	_, err := http.Post(e.CreateUrlEvent(eventKey), "application/json", bodyReader)
	return err
}

func (e *EventDispatcher) DispatchWithChannels(eventKey string, deviceID int, channelNumber int) error {
	bodyString := fmt.Sprintf(`{"deviceId": %d, "deviceByProps": [ "\"parentDevice\": %d", "\"channelNumber\": %d"]}`, deviceID, deviceID, channelNumber)
	body := []byte(bodyString)
	bodyReader := bytes.NewReader(body)
	_, err := http.Post(e.CreateUrlEvent(eventKey), "application/json",
		bodyReader)
	return err
}
func (e *EventDispatcher) DispatchWithChannelsAndOriginalValue(eventKey string, deviceID int, channelNumber int, originalValue string) error {
	bodyString := fmt.Sprintf(`{"deviceId": %d, "deviceByProps": [ "\"parentDevice\": %d", "\"channelNumber\": %d"], "originalValue": "%s"}`, deviceID, deviceID, channelNumber, originalValue)
	body := []byte(bodyString)
	bodyReader := bytes.NewReader(body)
	_, err := http.Post(e.CreateUrlEvent(eventKey), "application/json",
		bodyReader)
	return err
}

func (e *EventDispatcher) DispatchWithPersonIDAndOriginalValue(eventKey string, deviceID int, personId string, originalValue string) error {
	userType, userId, err := getUserTypeAndUserIdFromPersonID(personId)
	if err != nil {
		return err
	}
	bodyString := fmt.Sprintf(`{"deviceId": %d, "originalValue": "%s", "userType": %d, "userId": %d}`, deviceID, originalValue, userType, userId)
	body := []byte(bodyString)
	bodyReader := bytes.NewReader(body)
	_, err = http.Post(e.CreateUrlEvent(eventKey), "application/json",
		bodyReader)
	return err
}

func (e *EventDispatcher) DispatchWithPersonID(eventKey string, deviceID int, personId string) error {
	userType, userId, err := getUserTypeAndUserIdFromPersonID(personId)
	if err != nil {
		return err
	}
	bodyString := fmt.Sprintf(`{"deviceId": %d, "userType": %d, "userId": %d}`, deviceID, userType, userId)
	body := []byte(bodyString)
	bodyReader := bytes.NewReader(body)
	_, err = http.Post(e.CreateUrlEvent(eventKey), "application/json",
		bodyReader)
	return err
}

func (e *EventDispatcher) DispatchWithFile(eventKey string, deviceID int, file os.File) error {
	_, err := http.Post(e.CreateUrlEvent(eventKey), "application/json", &file)
	return err
}

func (e *EventDispatcher) CreateUrlEvent(eventKey string) string {
	var url string
	if e.port == nil {
		url = fmt.Sprintf("http://%s/v1/topologia/misc/%s", e.host, eventKey)
	} else {
		url = fmt.Sprintf("http://%s:%d/v1/topologia/misc/%s", e.host, *e.port, eventKey)
	}
	fmt.Println(url)
	return url
}

func getUserTypeAndUserIdFromPersonID(personId string) (userType int, userId int, err error) {
	if len(personId) < 3 {
		err = fmt.Errorf("invalid personId: %s", personId)
		return
	}

	userTypeString := personId[:2]
	userIdString := personId[2:]
	switch userTypeString {
	case "nu": // nu = Netsocs User
		userType = 1
	case "em": // em = Employee
		userType = 2
	case "vi": // vi = Visitor
		userType = 3
	default:
		err = fmt.Errorf("invalid personId: %s", personId)
		return
	}
	userId, err = strconv.Atoi(userIdString)
	return

}

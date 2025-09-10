package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/Netsocs-Team/driver.sdk_go/pkg/config"
	"github.com/Netsocs-Team/driver.sdk_go/pkg/objects"
	"github.com/Netsocs-Team/driver.sdk_go/pkg/tools"
	"github.com/go-resty/resty/v2"
)

type NetsocsDriverClient struct {
	driverKey           string
	driverHubHost       string
	isSSL               bool
	DriverName          string
	objectsRunner       objects.ObjectRunner
	siteID              string
	videoEngineID       string
	token               string
	driverID            string
	driverVersion       string
	driverDocumentation string
}

func (n *NetsocsDriverClient) SetVideoEngineID(videoEngineID string) {
	n.videoEngineID = videoEngineID
}

func (n *NetsocsDriverClient) SetSiteID(siteID string) {
	n.siteID = siteID
}

func (n *NetsocsDriverClient) SetToken(token string) {
	n.token = token
}

func (n *NetsocsDriverClient) SetDriverID(driverID string) {
	n.driverID = driverID
}

func (n *NetsocsDriverClient) GetToken() string {
	return n.token
}

func (n *NetsocsDriverClient) GetDriverID() string {
	return n.driverID
}

func (n *NetsocsDriverClient) GetSiteID() string {
	return n.siteID
}

func NewNetsocsDriverClient(driverKey string, driverHubHost string, isSSL bool) *NetsocsDriverClient {
	controller := objects.NewObjectController(driverHubHost, driverKey)
	go func() {
		err := controller.ListenActionRequests()
		if err != nil {
			panic(err)
		}
	}()
	runner := objects.NewObjectRunner(controller)
	client := &NetsocsDriverClient{
		driverKey:     driverKey,
		driverHubHost: driverHubHost,
		isSSL:         isSSL,
		objectsRunner: runner,
	}

	// If the events.json file exists, add the handler for the actionListenEvents
	// for create a default behavior for the actionListenEvents
	events, err := loadsEventsFromFile()
	if err == nil {
		err = client.AddConfigHandler(config.GET_EVENTS_AVAILABLE, func(valueMessage config.HandlerValue) (interface{}, error) {
			return events, nil
		})
		if err != nil {
			return nil
		}
	}

	return client
}

func New() (*NetsocsDriverClient, error) {

	fileData, err := tools.GetDriverNetsocsDotJsonContent("driver.netsocs.json")
	if err != nil {
		return nil, err
	}

	client := NewNetsocsDriverClient(fileData.DriverKey, fileData.DriverHubHost, false)
	client.DriverName = fileData.Name

	client.SetSiteID(fileData.SiteID)
	client.SetToken(fileData.Token)
	client.SetDriverID(fileData.DriverID)

	return client, nil
}

func (d *NetsocsDriverClient) GetChildren(parentId int) ([]Device, error) {
	url := d.buildURL(fmt.Sprintf("get_childs/%d", parentId))
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", d.driverKey)
	req.Header.Set("X-Auth-Token", d.token)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	var devices []Device
	if err := json.NewDecoder(res.Body).Decode(&devices); err != nil {
		return nil, err
	}

	for i := range devices {
		if devices[i].Params != nil && devices[i].Params["child_id"] != nil {
			childrenId := devices[i].Params["child_id"].(string)
			devices[i].ChildID = childrenId
		}
	}
	return devices, nil
}

func (d *NetsocsDriverClient) UploadFileAndGetURL(file *os.File) (string, error) {
	return tools.UploadFileAndGetURLWithReset(d.driverHubHost, d.driverKey, file)
}

func (d *NetsocsDriverClient) ListenConfig() error {

	return config.ListenConfig(d.driverHubHost, d.driverKey, d.siteID, d.token, d.driverID, func(videoEngineID string) {
		d.SetVideoEngineID(videoEngineID)
	}, d.driverVersion, d.driverDocumentation)
}

func (d *NetsocsDriverClient) SetDriverVersion(version string) {
	d.driverVersion = version
}

func (d *NetsocsDriverClient) SetDriverDocumentation(documentation string) {
	d.driverDocumentation = documentation
}

func (d *NetsocsDriverClient) AddConfigHandler(configKey config.NetsocsConfigKey, configHandler config.FuncConfigHandler) error {
	return config.AddConfigHandler(configKey, configHandler)
}

type LicenseMetadata struct {
	AccessControlDevices  int    `json:"accessControlDevices"`
	AlarmDevices          int    `json:"alarmDevices"`
	ClientName            string `json:"clientName"`
	CredentialModule      bool   `json:"credentialModule"`
	Drivers               int    `json:"drivers"`
	ExpirationDate        string `json:"expirationDate"`
	LidarDevices          int    `json:"lidarDevices"`
	OtherDevices          int    `json:"otherDevices"`
	Sites                 int    `json:"sites"`
	SupportExpiration     string `json:"supportExpiration"`
	VideoAnalyticsDevices int    `json:"videoAnalyticsDevices"`
	VideoChannels         int    `json:"videoChannels"`
	VisitModule           bool   `json:"visitModule"`
}

type LicenseInfo struct {
	Name             string          `json:"name"`
	Key              string          `json:"key"`
	Expiry           string          `json:"expiry"`
	Scheme           string          `json:"scheme"`
	RequireHeartbeat bool            `json:"requireHeartbeat"`
	LastValidated    string          `json:"lastValidated"`
	Created          string          `json:"created"`
	Updated          string          `json:"updated"`
	Metadata         LicenseMetadata `json:"metadata"`
}

type LicenseResponse struct {
	License struct {
		License LicenseInfo `json:"license"`
		Policy  interface{} `json:"policy"`
	} `json:"license"`
}

func (d *NetsocsDriverClient) GetLicense() (LicenseResponse, error) {
	resp, err := resty.New().R().SetHeader("X-Auth-Token", d.token).Get(d.driverHubHost + "/license")
	if err != nil {
		return LicenseResponse{}, err
	}

	var license LicenseResponse
	if err := json.Unmarshal(resp.Body(), &license); err != nil {
		return LicenseResponse{}, err
	}

	return license, nil
}

func (d *NetsocsDriverClient) buildURL(uri string) string {
	if d.isSSL {
		return fmt.Sprintf("https://%s/api/v1/%s", d.driverHubHost, uri)
	}
	return fmt.Sprintf("http://%s/api/v1/%s", d.driverHubHost, uri)
}

func (d *NetsocsDriverClient) GetDriverhubHost() string {
	return d.driverHubHost
}

func (c *NetsocsDriverClient) RegisterObject(obj objects.RegistrableObject) error {
	return c.objectsRunner.RegisterObject(obj)
}

func (c *NetsocsDriverClient) AddEventTypes(eventTypes []objects.EventType) error {
	err := c.objectsRunner.GetController().AddEventTypes(eventTypes)

	return err
}

func (c *NetsocsDriverClient) DispatchEvent(domain string, eventKey string, eventData objects.Event) (string, error) {
	req := objects.NewEventRequestBodySchema{}
	req.EventType = domain + "." + eventKey
	req.EventAdditionalProperties = eventData.Properties
	req.Images = eventData.ImageURLs
	req.VideoClips = eventData.VideoURLs

	for _, objID := range eventData.ObjectIDs {
		req.Rels = append(req.Rels, fmt.Sprintf("/objects/%s", objID))
	}

	resp, err := resty.New().R().SetHeader("X-Auth-Token", c.token).SetBody(req).Post(c.driverHubHost + "/objects/events")

	if err != nil {
		return "", err
	}
	return resp.String(), nil

}

func (c *NetsocsDriverClient) GetEvent(eventId string) (objects.EventRecord, error) {

	resp, err := resty.New().R().SetHeader("X-Auth-Token", c.token).Get(c.driverHubHost + "/objects/events/" + eventId)
	if err != nil {
		return objects.EventRecord{}, err
	}

	var event objects.EventRecord
	if err := json.Unmarshal(resp.Body(), &event); err != nil {
		return objects.EventRecord{}, err
	}

	return event, nil
}

func (c *NetsocsDriverClient) PatchEvent(eventId string, event objects.EventRecord) error {

	currentEvent, err := c.GetEvent(eventId)
	if err != nil {
		return err
	}

	if event.Images != nil {
		if len(event.Images) > 0 {
			currentEvent.Images = append(currentEvent.Images, event.Images...)
		} else {
			currentEvent.Images = event.Images
		}
	}

	if event.VideoClips != nil {
		if len(event.VideoClips) > 0 {

			currentEvent.VideoClips = append(currentEvent.VideoClips, event.VideoClips...)
		} else {
			currentEvent.VideoClips = event.VideoClips
		}
	}

	if event.EventAdditionalProperties != nil {
		if len(event.EventAdditionalProperties) > 0 {
			for key, value := range event.EventAdditionalProperties {
				currentEvent.EventAdditionalProperties[key] = value
			}
		} else {
			currentEvent.EventAdditionalProperties = event.EventAdditionalProperties
		}
	}

	resp, err := resty.New().R().SetHeader("X-Auth-Token", c.token).SetBody(currentEvent).Put(c.driverHubHost + "/objects/events/" + eventId)
	if err != nil {
		return err
	}

	if resp.IsError() {
		return errors.New(resp.String())
	}

	return nil
}

func (c *NetsocsDriverClient) SetObjectsBatchState(states []objects.ObjectStateChange) ([]objects.ChangeStateBatchResponse, error) {
	body := objects.ChangeStateBatchRequest{
		Changes: states,
	}
	resp, err := resty.New().R().SetHeader("X-Auth-Token", c.token).SetBody(body).Put(c.driverHubHost + "/objects/states-batch")
	if err != nil {
		return []objects.ChangeStateBatchResponse{}, err
	}

	if resp.IsError() {
		return []objects.ChangeStateBatchResponse{}, errors.New(resp.String())
	}
	var response []objects.ChangeStateBatchResponse
	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return []objects.ChangeStateBatchResponse{}, err
	}

	var errorResponse string
	for _, response := range response {
		if response.Error != "" {
			errorResponse += fmt.Sprintf("Error on object: %s: %s | ", response.ID, response.Error)
		}
	}
	if errorResponse != "" {
		return response, errors.New(errorResponse)
	}

	return response, nil
}

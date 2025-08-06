package client

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-resty/resty/v2"
)

func (d *NetsocsDriverClient) WriteLog(deviceId int, log string, params ...string) (*http.Response, error) {
	url := ""
	if d.isSSL {
		url = fmt.Sprintf("%s/devices/audit-logs", d.driverHubHost)
	} else {
		url = fmt.Sprintf("%s/devices/audit-logs", d.driverHubHost)
	}
	action := "NULL"
	if len(params) > 0 {
		action = params[0]
	}

	resp, err := resty.New().R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", d.token)).
		SetHeader("X-Auth-Token", d.token).
		SetBody(map[string]interface{}{
			"device_id": strconv.Itoa(deviceId),
			"message":   log,
			"action":    action,
		}).
		Post(url)

	if err != nil {
		return nil, err
	}
	return resp.RawResponse, nil
}

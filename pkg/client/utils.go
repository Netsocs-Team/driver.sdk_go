package client

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
)

type driverHubVersionResponse struct {
	GitCommitSha string `json:"git_commit_sha"`
	Version      string `json:"version"`
}

func (n *NetsocsDriverClient) checkVersion() error {
	resp, err := http.Get(n.buildURL("/api/v1/version"))

	if err != nil {
		return err
	}

	defer resp.Body.Close()
	data := driverHubVersionResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return err
	}

	if data.Version == "" {
		return fmt.Errorf("driver hub version must be greater than 2.0.0")
	}

	versionMajor := data.Version[0]

	if versionMajor < '2' {
		return fmt.Errorf("driver hub version must be greater than 2.0.0")
	}

	return nil
}

func (n *NetsocsDriverClient) RTSPToStreamID(rtsp string, name string) (string, error) {
	type responseSchema struct {
		StreamID string `json:"stream_id"`
	}
	responseBody := responseSchema{}
	client := resty.New()
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]string{"source": rtsp, "name": name}).
		Post(fmt.Sprintf("%s/objects/video-channels/encoded-sources", n.driverHubHost))

	if err != nil {
		return "", err
	}

	if err := json.Unmarshal(resp.Body(), &responseBody); err != nil {
		return "", err
	}

	return responseBody.StreamID, nil
}

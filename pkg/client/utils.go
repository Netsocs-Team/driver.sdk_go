package client

import (
	"encoding/json"
	"fmt"
	"net/http"
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

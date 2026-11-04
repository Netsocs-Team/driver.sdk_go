package tools

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/Netsocs-Team/driver.sdk_go/pkg/auth"
)

type FileSchema struct {
	DriverKey            string `json:"driver_key"`
	DriverHubHost        string `json:"driver_hub_host"`
	Version              string `json:"version"`
	Name                 string `json:"name"`
	DriverBinaryFilename string `json:"driver_binary_filename"`
	DocumentationURL     string `json:"documentation_url"`
	SiteID               string `json:"site_id"`
	SiteHost             string `json:"site_host"`
	Token                string `json:"token"`
	DriverID             string `json:"driver_id"`
	// Auth is written by SmartConnect when it enrolls the driver in Keycloak.
	// Absent: the driver uses the DriverKey, as always.
	Auth *auth.Config `json:"auth,omitempty"`
}

func GetDriverNetsocsDotJsonContent(path string) (*FileSchema, error) {
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var file FileSchema
	err = json.NewDecoder(f).Decode(&file)
	if err != nil {
		return nil, err
	}
	return &file, nil
}

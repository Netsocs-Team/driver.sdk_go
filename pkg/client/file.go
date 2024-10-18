package client

import (
	"encoding/json"
	"os"
)

type FileSchema struct {
	DriverKey                   string   `json:"driver_key"`
	DriverHubHost               string   `json:"driver_hub_host"`
	EventServerHost             string   `json:"event_server_host"`
	Version                     string   `json:"version"`
	Name                        string   `json:"name"`
	DriverBinaryFilename        string   `json:"driver_binary_filename"`
	ClusterMode                 bool     `json:"cluster_mode"`
	SettingsAvailable           []string `json:"settings_available"`
	LogLevel                    string   `json:"log_level"`
	DeviceModelsSupportedAll    bool     `json:"device_models_supported_all"`
	DeviceFirmwaresSupportedAll bool     `json:"device_firmwares_supported_all"`
	Comments                    string   `json:"comments"`
	DocumentationURL            string   `json:"documentation_url"`
}

func getDriverNetsocsDotJsonContent(path string) (*FileSchema, error) {
	f, err := os.Open(path)
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

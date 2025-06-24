package client

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type FileSchema struct {
	DriverKey            string `json:"driver_key"`
	DriverHubHost        string `json:"driver_hub_host"`
	Version              string `json:"version"`
	Name                 string `json:"name"`
	DriverBinaryFilename string `json:"driver_binary_filename"`
	DocumentationURL     string `json:"documentation_url"`
	SiteID               string `json:"site_id"`
}

func getDriverNetsocsDotJsonContent(path string) (*FileSchema, error) {
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

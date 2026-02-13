package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
)

// SnapshotUploadResponse is the response from the DriverHub /snapshots/upload endpoint.
type SnapshotUploadResponse struct {
	Filename string `json:"filename"`
	URL      string `json:"url"`
	Path     string `json:"path"`
}

// UploadSnapshot uploads image data to the DriverHub snapshots endpoint.
// r is the image content (e.g. *os.File, *bytes.Reader, HTTP response body).
// filename is used as the multipart file name and, when customName is empty, as the stored name (e.g. "snapshot.jpg").
// customName is optional; if empty, filename is used. The backend validates image type (.jpg, .jpeg, .png).
// Returns the response with filename, url and path (e.g. "/public/filename.jpg").
func UploadSnapshot(driverHubHost, driverKey string, r io.Reader, filename string, customName string) (SnapshotUploadResponse, error) {
	var empty SnapshotUploadResponse
	if !strings.HasPrefix(driverHubHost, "http://") && !strings.HasPrefix(driverHubHost, "https://") {
		driverHubHost = fmt.Sprintf("http://%s", driverHubHost)
	}
	url := strings.TrimSuffix(driverHubHost, "/") + "/snapshots/upload"

	fileContent, err := io.ReadAll(r)
	if err != nil {
		return empty, fmt.Errorf("error reading content: %w", err)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	nameInForm := filepath.Base(filename)
	if nameInForm == "" || nameInForm == "." {
		nameInForm = "snapshot.jpg"
	}
	part, err := writer.CreateFormFile("file", nameInForm)
	if err != nil {
		return empty, fmt.Errorf("error creating form file: %w", err)
	}
	if _, err = part.Write(fileContent); err != nil {
		return empty, fmt.Errorf("error writing file content to form: %w", err)
	}
	if customName != "" {
		_ = writer.WriteField("name", customName)
	}
	contentType := writer.FormDataContentType()
	if err = writer.Close(); err != nil {
		return empty, fmt.Errorf("error closing multipart writer: %w", err)
	}

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return empty, fmt.Errorf("error creating HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", driverKey)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return empty, fmt.Errorf("error making HTTP request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		content, _ := io.ReadAll(res.Body)
		return empty, fmt.Errorf("HTTP error %d: %s", res.StatusCode, string(content))
	}

	content, err := io.ReadAll(res.Body)
	if err != nil {
		return empty, fmt.Errorf("error reading response body: %w", err)
	}

	var response SnapshotUploadResponse
	if err := json.Unmarshal(content, &response); err != nil {
		return empty, fmt.Errorf("error unmarshaling response: %w", err)
	}
	return response, nil
}

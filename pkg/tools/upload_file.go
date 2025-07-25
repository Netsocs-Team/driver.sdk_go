package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type uploadFileResponse struct {
	Filename string `json:"filename"`
}

func UploadFileAndGetURL(driverHubHost string, driverKey string, file *os.File) (string, error) {
	body := &bytes.Buffer{}
	if !strings.HasPrefix(driverHubHost, "http://") && !strings.HasPrefix(driverHubHost, "https://") {
		driverHubHost = fmt.Sprintf("http://%s", driverHubHost)
	}
	url := fmt.Sprintf("%s/api/v1/upload", driverHubHost)

	// Create a form file
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(file.Name()))
	if err != nil {
		return "", fmt.Errorf("error creating form file: %w", err)
	}

	// Copy file content to the form part
	_, err = io.Copy(part, file)
	if err != nil {
		return "", fmt.Errorf("error copying file content: %w", err)
	}

	contentType := writer.FormDataContentType()
	writer.Close()

	// Create HTTP request with authentication
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return "", fmt.Errorf("error creating HTTP request: %w", err)
	}

	// Add authentication headers
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", driverKey)

	// Make the request
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making HTTP request: %w", err)
	}
	defer res.Body.Close()

	// Check HTTP status code
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		content, _ := io.ReadAll(res.Body)
		return "", fmt.Errorf("HTTP error %d: %s", res.StatusCode, string(content))
	}

	content, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	var response uploadFileResponse
	if err := json.Unmarshal(content, &response); err != nil {
		return "", fmt.Errorf("error unmarshaling response: %w", err)
	}

	return fmt.Sprintf("%s/public/%s", driverHubHost, response.Filename), nil
}

// UploadFileAndGetURLWithReset uploads a file and resets its position to the beginning
// This is useful when you need to reuse the file after upload
func UploadFileAndGetURLWithReset(driverHubHost string, driverKey string, file *os.File) (string, error) {
	// Get current position
	currentPos, err := file.Seek(0, io.SeekCurrent)
	if err != nil {
		return "", fmt.Errorf("error getting current file position: %w", err)
	}

	// Read the entire file content into a buffer
	fileInfo, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("error getting file info: %w", err)
	}

	// Reset to beginning to read all content
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return "", fmt.Errorf("error seeking to file start: %w", err)
	}

	fileContent := make([]byte, fileInfo.Size())
	_, err = io.ReadFull(file, fileContent)
	if err != nil {
		return "", fmt.Errorf("error reading file content: %w", err)
	}

	// Reset to original position
	_, err = file.Seek(currentPos, io.SeekStart)
	if err != nil {
		return "", fmt.Errorf("error resetting file position: %w", err)
	}

	// Create a buffer with the file content
	body := &bytes.Buffer{}
	if !strings.HasPrefix(driverHubHost, "http://") && !strings.HasPrefix(driverHubHost, "https://") {
		driverHubHost = fmt.Sprintf("http://%s", driverHubHost)
	}
	url := fmt.Sprintf("%s/api/v1/upload", driverHubHost)

	// Create a form file
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(file.Name()))
	if err != nil {
		return "", fmt.Errorf("error creating form file: %w", err)
	}

	// Copy file content from buffer to the form part
	_, err = part.Write(fileContent)
	if err != nil {
		return "", fmt.Errorf("error writing file content to form: %w", err)
	}

	contentType := writer.FormDataContentType()
	writer.Close()

	// Create HTTP request with authentication
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return "", fmt.Errorf("error creating HTTP request: %w", err)
	}

	// Add authentication headers
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", driverKey)

	// Make the request
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making HTTP request: %w", err)
	}
	defer res.Body.Close()

	// Check HTTP status code
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		content, _ := io.ReadAll(res.Body)
		return "", fmt.Errorf("HTTP error %d: %s", res.StatusCode, string(content))
	}

	content, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	var response uploadFileResponse
	if err := json.Unmarshal(content, &response); err != nil {
		return "", fmt.Errorf("error unmarshaling response: %w", err)
	}

	return fmt.Sprintf("%s/public/%s", driverHubHost, response.Filename), nil
}

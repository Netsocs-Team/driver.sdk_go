package tools

import (
	"os"
	"strings"
	"testing"
)

func TestUploadFileAndGetURL(t *testing.T) {
	// Create a temporary test file
	testContent := "This is a test file content"
	tmpfile, err := os.CreateTemp("", "test_upload_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	// Write test content to the file
	if _, err := tmpfile.Write([]byte(testContent)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	// Close the file to ensure it's written to disk
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	// Reopen the file for reading
	file, err := os.Open(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to open temp file: %v", err)
	}
	defer file.Close()

	// Test with invalid host (should fail gracefully)
	_, err = UploadFileAndGetURL("invalid-host", "test-key", file)
	if err == nil {
		t.Error("Expected error for invalid host, got nil")
	}

	// Verify the error message contains useful information
	if !strings.Contains(err.Error(), "error making HTTP request") {
		t.Errorf("Expected error to contain 'error making HTTP request', got: %s", err.Error())
	}
}

func TestUploadFileAndGetURLWithReset(t *testing.T) {
	// Create a temporary test file
	testContent := "This is a test file content for reset test"
	tmpfile, err := os.CreateTemp("", "test_upload_reset_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	// Write test content to the file
	if _, err := tmpfile.Write([]byte(testContent)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	// Close the file to ensure it's written to disk
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	// Reopen the file for reading
	file, err := os.Open(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to open temp file: %v", err)
	}
	defer file.Close()

	// Read some content to move the file pointer
	buffer := make([]byte, 10)
	_, err = file.Read(buffer)
	if err != nil {
		t.Fatalf("Failed to read from file: %v", err)
	}

	// Get current position
	initialPos, err := file.Seek(0, 1) // SeekCurrent
	if err != nil {
		t.Fatalf("Failed to get current position: %v", err)
	}

	// Test with invalid host (should fail gracefully)
	_, err = UploadFileAndGetURLWithReset("invalid-host", "test-key", file)
	if err == nil {
		t.Error("Expected error for invalid host, got nil")
	}

	// Verify the error message contains useful information
	if !strings.Contains(err.Error(), "error making HTTP request") {
		t.Errorf("Expected error to contain 'error making HTTP request', got: %s", err.Error())
	}

	// Check that file position was reset
	finalPos, err := file.Seek(0, 1) // SeekCurrent
	if err != nil {
		t.Fatalf("Failed to get final position: %v", err)
	}

	if finalPos != initialPos {
		t.Errorf("File position was not reset correctly. Expected %d, got %d", initialPos, finalPos)
	}
}

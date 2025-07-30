package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServer_Start(t *testing.T) {
	dir := "../utils" // Use utils dir for test files
	port := 8081
	s := NewServer(dir, port)

	// Create a test server using the handler
	fs := http.FileServer(http.Dir(s.Dir))
	ts := httptest.NewServer(fs)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/utils.go")
	if err != nil {
		t.Fatalf("Failed to GET file: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestServer_PartialDownload(t *testing.T) {
	dir := "../utils" // Use utils dir for test files
	port := 8082
	s := NewServer(dir, port)

	// Create a test server using the handler
	fs := http.FileServer(http.Dir(s.Dir))
	ts := httptest.NewServer(fs)
	defer ts.Close()

	req, err := http.NewRequest("GET", ts.URL+"/utils.go", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Range", "bytes=0-99")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to GET file with range: %v", err)
	}
	if resp.StatusCode != http.StatusPartialContent {
		t.Errorf("Expected status 206, got %d", resp.StatusCode)
	}

	if resp.Header.Get("Content-Range") == "" {
		t.Error("Expected Content-Range header, got none")
	}
}

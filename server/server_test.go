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

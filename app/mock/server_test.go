package mock

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHealthEndpoint(t *testing.T) {
	mockServer := NewMockServer(":9090")

	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(mockServer.handleHealth)
	handler.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check Content-Type
	expected := "application/json"
	if ctype := rr.Header().Get("Content-Type"); ctype != expected {
		t.Errorf("handler returned wrong content type: got %v want %v", ctype, expected)
	}

	// Check X-Mock-Server header
	expectedHeader := "nproxy-mock"
	if mockHeader := rr.Header().Get("X-Mock-Server"); mockHeader != expectedHeader {
		t.Errorf("handler returned wrong X-Mock-Server header: got %v want %v", mockHeader, expectedHeader)
	}

	// Check response body
	var response HealthResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Could not parse response JSON: %v", err)
	}

	if response.Status != "healthy" {
		t.Errorf("Expected status 'healthy', got %v", response.Status)
	}

	if response.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got %v", response.Version)
	}

	// Check if timestamp is recent
	if time.Since(response.Timestamp) > time.Minute {
		t.Errorf("Timestamp is too old: %v", response.Timestamp)
	}
}

func TestUsersEndpoint(t *testing.T) {
	mockServer := NewMockServer(":9090")

	req, err := http.NewRequest("GET", "/api/users", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(mockServer.handleUsers)
	handler.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check response body
	var users []map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &users); err != nil {
		t.Errorf("Could not parse response JSON: %v", err)
	}

	if len(users) != 3 {
		t.Errorf("Expected 3 users, got %d", len(users))
	}

	// Check first user
	if users[0]["name"] != "Alice" {
		t.Errorf("Expected first user name to be 'Alice', got %v", users[0]["name"])
	}
}

func TestUsersEndpointMethodNotAllowed(t *testing.T) {
	mockServer := NewMockServer(":9090")

	req, err := http.NewRequest("POST", "/api/users", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(mockServer.handleUsers)
	handler.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusMethodNotAllowed)
	}
}

func TestEchoEndpoint(t *testing.T) {
	mockServer := NewMockServer(":9090")

	testBody := `{"test": "data"}`
	req, err := http.NewRequest("POST", "/api/echo", bytes.NewBufferString(testBody))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(mockServer.handleEcho)
	handler.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check response body
	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Could not parse response JSON: %v", err)
	}

	if response["method"] != "POST" {
		t.Errorf("Expected method 'POST', got %v", response["method"])
	}

	if response["body"] != testBody {
		t.Errorf("Expected body '%s', got %v", testBody, response["body"])
	}
}

func TestDefaultEndpoint(t *testing.T) {
	mockServer := NewMockServer(":9090")

	req, err := http.NewRequest("GET", "/unknown", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(mockServer.handleDefault)
	handler.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check response body
	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Could not parse response JSON: %v", err)
	}

	if response["message"] != "Mock server is running" {
		t.Errorf("Expected message 'Mock server is running', got %v", response["message"])
	}

	if response["path"] != "/unknown" {
		t.Errorf("Expected path '/unknown', got %v", response["path"])
	}
}

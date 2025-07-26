package mock

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"
)

// HealthResponse is the response structure for health check
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
	Version   string    `json:"version"`
}

// MockServer is the mock server structure
type MockServer struct {
	addr string
}

// NewMockServer creates a new mock server
func NewMockServer(addr string) *MockServer {
	return &MockServer{
		addr: addr,
	}
}

// Start starts the mock server
func (m *MockServer) Start() error {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", m.handleHealth)

	// Other test endpoints
	mux.HandleFunc("/api/users", m.handleUsers)
	mux.HandleFunc("/api/echo", m.handleEcho)
	mux.HandleFunc("/", m.handleDefault)

	log.Printf("Starting mock server on %s", m.addr)
	log.Printf("Available endpoints:")
	log.Printf("  GET %s/health - Health check", m.addr)
	log.Printf("  GET %s/api/users - Mock users API", m.addr)
	log.Printf("  POST %s/api/echo - Echo request body", m.addr)

	return http.ListenAndServe(m.addr, mux)
}

// handleHealth handles health check
func (m *MockServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	log.Printf("Health check request from %s", r.RemoteAddr)

	// Log request headers (check headers via proxy)
	log.Printf("Request headers:")
	for key, values := range r.Header {
		log.Printf("  %s: %v", key, values)
	}

	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Message:   "Mock server is running properly",
		Version:   "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Mock-Server", "nproxy-mock")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding health response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("Health check response sent successfully")
}

// handleUsers handles mock user list API
func (m *MockServer) handleUsers(w http.ResponseWriter, r *http.Request) {
	log.Printf("Users API request from %s", r.RemoteAddr)

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	users := []map[string]interface{}{
		{"id": 1, "name": "Alice", "email": "alice@example.com"},
		{"id": 2, "name": "Bob", "email": "bob@example.com"},
		{"id": 3, "name": "Charlie", "email": "charlie@example.com"},
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Mock-Server", "nproxy-mock")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(users); err != nil {
		log.Printf("Error encoding users response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("Users API response sent successfully")
}

// handleEcho handles echo API that returns request body as-is
func (m *MockServer) handleEcho(w http.ResponseWriter, r *http.Request) {
	log.Printf("Echo API request from %s", r.RemoteAddr)

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Read request body
	var body []byte
	var err error
	if r.ContentLength > 0 {
		body = make([]byte, r.ContentLength)
		_, err = r.Body.Read(body)
		if err != nil && err.Error() != "EOF" {
			log.Printf("Error reading request body: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
	} else {
		// Use io.ReadAll for unknown content length or when ContentLength is not set properly
		body, err = io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading request body: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
	}
	defer r.Body.Close()

	response := map[string]interface{}{
		"method":    r.Method,
		"url":       r.URL.String(),
		"headers":   r.Header,
		"body":      string(body),
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Mock-Server", "nproxy-mock")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding echo response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("Echo API response sent successfully")
}

// handleDefault is the default handler
func (m *MockServer) handleDefault(w http.ResponseWriter, r *http.Request) {
	log.Printf("Default handler request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

	response := map[string]interface{}{
		"message":   "Mock server is running",
		"path":      r.URL.Path,
		"method":    r.Method,
		"timestamp": time.Now(),
		"endpoints": []string{
			"/health",
			"/api/users",
			"/api/echo",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Mock-Server", "nproxy-mock")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding default response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// Start is a function to start mock server standalone
func Start(addr string) error {
	server := NewMockServer(addr)
	return server.Start()
}

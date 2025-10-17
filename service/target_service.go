package service

import (
	"encoding/json"
	"fmt"
	"incident-ai/models"
	"log"
	"net/http"
	"sync"
	"time"
)

// TargetService represents a service that can experience incidents
type TargetService struct {
	port          string
	isHealthy     bool
	isRunning     bool
	config        map[string]string
	mu            sync.RWMutex
	server        *http.Server
	errorLogs     []string
	maxLogs       int
}

// NewTargetService creates a new target service
func NewTargetService(port string) *TargetService {
	return &TargetService{
		port:      port,
		isHealthy: true,
		isRunning: false,
		config: map[string]string{
			"database_url": "localhost:5432",
			"timeout":      "30s",
			"max_retries":  "3",
		},
		errorLogs: make([]string, 0),
		maxLogs:   50,
	}
}

// Start starts the target service
func (ts *TargetService) Start() error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.isRunning {
		return fmt.Errorf("service already running")
	}

	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", ts.handleHealth)

	// Trigger incident endpoint
	mux.HandleFunc("/trigger-incident", ts.handleTriggerIncident)

	// Normal API endpoint
	mux.HandleFunc("/api/data", ts.handleAPI)

	// Status endpoint
	mux.HandleFunc("/status", ts.handleStatus)

	ts.server = &http.Server{
		Addr:    ":" + ts.port,
		Handler: mux,
	}

	ts.isRunning = true
	ts.isHealthy = true

	go func() {
		log.Printf("[TARGET SERVICE] Starting on port %s\n", ts.port)
		if err := ts.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			ts.addLog(fmt.Sprintf("Server error: %v", err))
			log.Printf("[TARGET SERVICE] Error: %v\n", err)
		}
	}()

	time.Sleep(500 * time.Millisecond) // Give server time to start
	return nil
}

// Stop stops the target service
func (ts *TargetService) Stop() error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if !ts.isRunning {
		return fmt.Errorf("service not running")
	}

	ts.isRunning = false
	ts.isHealthy = false

	if ts.server != nil {
		return ts.server.Close()
	}
	return nil
}

// IsHealthy returns the health status
func (ts *TargetService) IsHealthy() bool {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.isHealthy && ts.isRunning
}

// GetLogs returns recent error logs
func (ts *TargetService) GetLogs() []string {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	logs := make([]string, len(ts.errorLogs))
	copy(logs, ts.errorLogs)
	return logs
}

// GetConfig returns current configuration
func (ts *TargetService) GetConfig() map[string]string {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	config := make(map[string]string)
	for k, v := range ts.config {
		config[k] = v
	}
	return config
}

// SetConfig updates configuration
func (ts *TargetService) SetConfig(key, value string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.config[key] = value
}

// Restart restarts the service
func (ts *TargetService) Restart() error {
	log.Println("[TARGET SERVICE] Restarting...")

	if err := ts.Stop(); err != nil && ts.isRunning {
		return err
	}

	time.Sleep(1 * time.Second)

	return ts.Start()
}

func (ts *TargetService) addLog(message string) {
	ts.errorLogs = append(ts.errorLogs, fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), message))
	if len(ts.errorLogs) > ts.maxLogs {
		ts.errorLogs = ts.errorLogs[1:]
	}
}

// HTTP Handlers

func (ts *TargetService) handleHealth(w http.ResponseWriter, r *http.Request) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	status := models.HealthStatus{
		Healthy:   ts.isHealthy,
		Timestamp: time.Now(),
		Message:   "Service operational",
	}

	w.Header().Set("Content-Type", "application/json")

	if !ts.isHealthy {
		status.Message = "Service unhealthy"
		status.StatusCode = http.StatusServiceUnavailable
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		status.StatusCode = http.StatusOK
		w.WriteHeader(http.StatusOK)
	}

	json.NewEncoder(w).Encode(status)
}

func (ts *TargetService) handleTriggerIncident(w http.ResponseWriter, r *http.Request) {
	incidentType := r.URL.Query().Get("type")

	log.Printf("[TARGET SERVICE] Triggering incident: %s\n", incidentType)

	ts.mu.Lock()
	defer ts.mu.Unlock()

	switch incidentType {
	case "crash", "SERVICE_DOWN":
		ts.isHealthy = false
		ts.addLog("Service crashed - simulated failure")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Incident triggered: SERVICE_DOWN\n")

	case "config", "CONFIG_ERROR":
		ts.config["database_url"] = "invalid::url::format"
		ts.config["timeout"] = "not-a-number"
		ts.isHealthy = false
		ts.addLog("Configuration corrupted - invalid values detected")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Incident triggered: CONFIG_ERROR\n")

	case "resource", "RESOURCE_EXHAUSTION":
		ts.isHealthy = false
		ts.addLog("Resource exhaustion - port blocked or memory full")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Incident triggered: RESOURCE_EXHAUSTION\n")

	case "dependency", "DEPENDENCY_FAILURE":
		ts.config["database_url"] = "unreachable-host:9999"
		ts.isHealthy = false
		ts.addLog("Database connection failed - unable to reach host")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Incident triggered: DEPENDENCY_FAILURE\n")

	default:
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Unknown incident type: %s\n", incidentType)
		fmt.Fprintf(w, "Valid types: crash, config, resource, dependency\n")
		return
	}
}

func (ts *TargetService) handleAPI(w http.ResponseWriter, r *http.Request) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	if !ts.isHealthy {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "service unavailable"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"data":   "Sample API response",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func (ts *TargetService) handleStatus(w http.ResponseWriter, r *http.Request) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"running":    ts.isRunning,
		"healthy":    ts.isHealthy,
		"config":     ts.config,
		"recent_logs": ts.errorLogs,
	})
}

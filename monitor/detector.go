package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"incident-ai/models"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// IncidentDetector monitors services and detects incidents
type IncidentDetector struct {
	serviceURL      string
	checkInterval   time.Duration
	incidentChannel chan *models.Incident
	stopChannel     chan bool
	isRunning       bool
}

// NewIncidentDetector creates a new incident detector
func NewIncidentDetector(serviceURL string, checkInterval time.Duration) *IncidentDetector {
	return &IncidentDetector{
		serviceURL:      serviceURL,
		checkInterval:   checkInterval,
		incidentChannel: make(chan *models.Incident, 10),
		stopChannel:     make(chan bool),
		isRunning:       false,
	}
}

// Start begins monitoring
func (id *IncidentDetector) Start(ctx context.Context) {
	if id.isRunning {
		log.Println("[MONITOR] Already running")
		return
	}

	id.isRunning = true
	log.Printf("[MONITOR] Started monitoring %s (interval: %v)\n", id.serviceURL, id.checkInterval)

	go id.monitorLoop(ctx)
}

// Stop stops monitoring
func (id *IncidentDetector) Stop() {
	if !id.isRunning {
		return
	}

	log.Println("[MONITOR] Stopping...")
	id.stopChannel <- true
	id.isRunning = false
}

// GetIncidentChannel returns the channel where incidents are published
func (id *IncidentDetector) GetIncidentChannel() <-chan *models.Incident {
	return id.incidentChannel
}

func (id *IncidentDetector) monitorLoop(ctx context.Context) {
	ticker := time.NewTicker(id.checkInterval)
	defer ticker.Stop()

	previousHealthy := true

	for {
		select {
		case <-ctx.Done():
			log.Println("[MONITOR] Context cancelled")
			return

		case <-id.stopChannel:
			log.Println("[MONITOR] Stopped")
			return

		case <-ticker.C:
			health := id.checkHealth()

			// Only trigger incident on transition from healthy to unhealthy
			if previousHealthy && !health.Healthy {
				log.Println("[MONITOR] ⚠️  Health check FAILED - Incident detected!")
				incident := id.createIncident(health)
				id.incidentChannel <- incident
			} else if !previousHealthy && health.Healthy {
				log.Println("[MONITOR] ✓ Health check PASSED - Service recovered")
			}

			previousHealthy = health.Healthy
		}
	}
}

func (id *IncidentDetector) checkHealth() models.HealthStatus {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(id.serviceURL + "/health")
	if err != nil {
		return models.HealthStatus{
			Healthy:   false,
			Timestamp: time.Now(),
			Message:   fmt.Sprintf("Health check failed: %v", err),
			StatusCode: 0,
		}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var healthStatus models.HealthStatus
	if err := json.Unmarshal(body, &healthStatus); err != nil {
		return models.HealthStatus{
			Healthy:   false,
			Timestamp: time.Now(),
			Message:   "Failed to parse health response",
			StatusCode: resp.StatusCode,
		}
	}

	healthStatus.StatusCode = resp.StatusCode
	return healthStatus
}

func (id *IncidentDetector) createIncident(health models.HealthStatus) *models.Incident {
	// Determine incident type and gather symptoms
	incidentType, symptoms := id.analyzeSymptoms(health)

	// Fetch logs from the service
	logs := id.fetchLogs()

	incident := &models.Incident{
		ID:         uuid.New().String(),
		Type:       incidentType,
		Status:     models.StatusDetected,
		DetectedAt: time.Now(),
		Symptoms:   symptoms,
		Logs:       logs,
		UsedCachedFix: false,
	}

	return incident
}

func (id *IncidentDetector) analyzeSymptoms(health models.HealthStatus) (models.IncidentType, []string) {
	symptoms := []string{
		fmt.Sprintf("Health check returned status code: %d", health.StatusCode),
		health.Message,
	}

	// Get current service status for more context
	status := id.fetchServiceStatus()

	if config, ok := status["config"].(map[string]interface{}); ok {
		// Check for config issues
		if dbURL, exists := config["database_url"]; exists {
			if str, ok := dbURL.(string); ok && (str == "invalid::url::format" || str == "") {
				symptoms = append(symptoms, "Invalid database URL configuration detected")
				return models.ConfigError, symptoms
			}
			if str, ok := dbURL.(string); ok && str == "unreachable-host:9999" {
				symptoms = append(symptoms, "Database host unreachable")
				return models.DependencyFailure, symptoms
			}
		}
		if timeout, exists := config["timeout"]; exists {
			if str, ok := timeout.(string); ok && str == "not-a-number" {
				symptoms = append(symptoms, "Invalid timeout configuration detected")
				return models.ConfigError, symptoms
			}
		}
	}

	// Check if service is not running at all
	if running, ok := status["running"].(bool); ok && !running {
		symptoms = append(symptoms, "Service process not running")
		return models.ServiceDown, symptoms
	}

	// Check logs for resource issues
	if logs, ok := status["recent_logs"].([]interface{}); ok && len(logs) > 0 {
		for _, logEntry := range logs {
			if str, ok := logEntry.(string); ok {
				if contains(str, "resource") || contains(str, "port blocked") || contains(str, "memory") {
					symptoms = append(symptoms, "Resource exhaustion detected in logs")
					return models.ResourceExhaustion, symptoms
				}
			}
		}
	}

	// Default to service down
	symptoms = append(symptoms, "Service health check failing")
	return models.ServiceDown, symptoms
}

func (id *IncidentDetector) fetchLogs() []string {
	status := id.fetchServiceStatus()

	if logs, ok := status["recent_logs"].([]interface{}); ok {
		strLogs := make([]string, 0, len(logs))
		for _, log := range logs {
			if str, ok := log.(string); ok {
				strLogs = append(strLogs, str)
			}
		}
		return strLogs
	}

	return []string{}
}

func (id *IncidentDetector) fetchServiceStatus() map[string]interface{} {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(id.serviceURL + "/status")
	if err != nil {
		return map[string]interface{}{}
	}
	defer resp.Body.Close()

	var status map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return map[string]interface{}{}
	}

	return status
}

// VerifyResolution checks if an incident has been resolved
func (id *IncidentDetector) VerifyResolution() bool {
	health := id.checkHealth()
	return health.Healthy
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
		len(s) > len(substr) && hasSubstring(s, substr)))
}

func hasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

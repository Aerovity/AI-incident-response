package models

import "time"

// IncidentType represents the type of incident
type IncidentType string

const (
	ServiceDown        IncidentType = "SERVICE_DOWN"
	ConfigError        IncidentType = "CONFIG_ERROR"
	ResourceExhaustion IncidentType = "RESOURCE_EXHAUSTION"
	DependencyFailure  IncidentType = "DEPENDENCY_FAILURE"
	DatabaseError      IncidentType = "DATABASE_ERROR"
	HighLatency        IncidentType = "HIGH_LATENCY"
	MemoryLeak         IncidentType = "MEMORY_LEAK"
	SecurityBreach     IncidentType = "SECURITY_BREACH"
	NetworkPartition   IncidentType = "NETWORK_PARTITION"
	DiskFull           IncidentType = "DISK_FULL"
)

// IncidentStatus represents the current state of an incident
type IncidentStatus string

const (
	StatusDetected  IncidentStatus = "DETECTED"
	StatusAnalyzing IncidentStatus = "ANALYZING"
	StatusFixing    IncidentStatus = "FIXING"
	StatusResolved  IncidentStatus = "RESOLVED"
	StatusFailed    IncidentStatus = "FAILED"
)

// IncidentPriority represents the priority level of an incident
type IncidentPriority string

const (
	PriorityCritical IncidentPriority = "CRITICAL"
	PriorityHigh     IncidentPriority = "HIGH"
	PriorityMedium   IncidentPriority = "MEDIUM"
	PriorityLow      IncidentPriority = "LOW"
)

// Incident represents a detected system incident
type Incident struct {
	ID            string           `json:"id"`
	Type          IncidentType     `json:"type"`
	Status        IncidentStatus   `json:"status"`
	Priority      IncidentPriority `json:"priority"`
	DetectedAt    time.Time        `json:"detected_at"`
	ResolvedAt    *time.Time       `json:"resolved_at,omitempty"`
	Symptoms      []string         `json:"symptoms"`
	Logs          []string         `json:"logs"`
	Diagnosis     string           `json:"diagnosis,omitempty"`
	Resolution    *Resolution      `json:"resolution,omitempty"`
	UsedCachedFix bool             `json:"used_cached_fix"`
	RollbackData  *RollbackData    `json:"rollback_data,omitempty"`
}

// Resolution represents how an incident was fixed
type Resolution struct {
	FixType     string   `json:"fix_type"`     // "code", "config", "restart"
	Description string   `json:"description"`
	Steps       []string `json:"steps"`
	Code        string   `json:"code,omitempty"`
	Success     bool     `json:"success"`
}

// AIResponse represents the response from the AI
type AIResponse struct {
	Diagnosis   string   `json:"diagnosis"`
	FixType     string   `json:"fix_type"`
	FixSteps    []string `json:"fix_steps"`
	Code        string   `json:"code,omitempty"`
	Confidence  float64  `json:"confidence,omitempty"`
}

// HealthStatus represents the health of a service
type HealthStatus struct {
	Healthy   bool      `json:"healthy"`
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
	StatusCode int      `json:"status_code,omitempty"`
}

// RollbackData stores information needed to rollback a change
type RollbackData struct {
	Timestamp      time.Time              `json:"timestamp"`
	PreviousState  map[string]interface{} `json:"previous_state"`
	ConfigBackup   string                 `json:"config_backup,omitempty"`
	ServiceState   string                 `json:"service_state"`
	CanRollback    bool                   `json:"can_rollback"`
	RollbackSteps  []string               `json:"rollback_steps,omitempty"`
}

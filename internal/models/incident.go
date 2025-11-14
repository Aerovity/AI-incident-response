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
	StatusOpen          IncidentStatus = "OPEN"
	StatusInvestigating IncidentStatus = "INVESTIGATING"
	StatusIdentified    IncidentStatus = "IDENTIFIED"
	StatusMonitoring    IncidentStatus = "MONITORING"
	StatusResolved      IncidentStatus = "RESOLVED"
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
	ID          string           `json:"id"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Type        IncidentType     `json:"type"`
	Status      IncidentStatus   `json:"status"`
	Priority    IncidentPriority `json:"priority"`

	// Multi-tenancy
	OrganizationID string `json:"organization_id"`

	// Assignment
	ReporterID   string  `json:"reporter_id"`
	AssignedToID *string `json:"assigned_to_id,omitempty"`
	TeamID       *string `json:"team_id,omitempty"`

	// Timeline
	DetectedAt     time.Time  `json:"detected_at"`
	AcknowledgedAt *time.Time `json:"acknowledged_at,omitempty"`
	ResolvedAt     *time.Time `json:"resolved_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`

	// AI & Remediation
	AIAnalysis     *AIAnalysis `json:"ai_analysis,omitempty"`
	RemediationLog []string    `json:"remediation_log,omitempty"`

	// Slack Integration
	SlackChannelID string `json:"slack_channel_id,omitempty"`

	// Updates/Comments
	Updates []IncidentUpdate `json:"updates,omitempty"`
}

// IncidentUpdate represents a comment or update on an incident
type IncidentUpdate struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// AIAnalysis represents AI analysis of an incident
type AIAnalysis struct {
	Diagnosis  string    `json:"diagnosis"`
	FixType    string    `json:"fix_type"`
	Steps      []string  `json:"steps"`
	Confidence float64   `json:"confidence"`
	AnalyzedAt time.Time `json:"analyzed_at"`
}

// CreateIncidentRequest represents the request to create a new incident
type CreateIncidentRequest struct {
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Type        IncidentType     `json:"type"`
	Priority    IncidentPriority `json:"priority"`
}

// UpdateIncidentRequest represents the request to update an incident
type UpdateIncidentRequest struct {
	Title        *string           `json:"title,omitempty"`
	Description  *string           `json:"description,omitempty"`
	Status       *IncidentStatus   `json:"status,omitempty"`
	Priority     *IncidentPriority `json:"priority,omitempty"`
	AssignedToID *string           `json:"assigned_to_id,omitempty"`
	TeamID       *string           `json:"team_id,omitempty"`
}

// AddUpdateRequest represents a request to add an update/comment
type AddUpdateRequest struct {
	Message string `json:"message"`
}

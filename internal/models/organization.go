package models

import "time"

// Organization represents a tenant organization
type Organization struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Slug      string     `json:"slug"` // unique identifier
	Plan      string     `json:"plan"` // free, starter, pro
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`

	// Settings
	AutoRemediation bool         `json:"auto_remediation"`
	SlackWebhookURL string       `json:"slack_webhook_url,omitempty"`
	SMTPConfig      *SMTPConfig  `json:"smtp_config,omitempty"`
}

// SMTPConfig holds SMTP configuration for email notifications
type SMTPConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	From     string `json:"from"`
}

package models

import "time"

// Team represents a team within an organization
type Team struct {
	ID             string           `json:"id"`
	Name           string           `json:"name"`
	Description    string           `json:"description,omitempty"`
	OrganizationID string           `json:"organization_id"`
	MemberIDs      []string         `json:"member_ids"`
	OnCallSchedule *OnCallSchedule  `json:"on_call_schedule,omitempty"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
}

// OnCallSchedule represents the on-call rotation for a team
type OnCallSchedule struct {
	CurrentUserID string    `json:"current_user_id"`
	Rotation      []string  `json:"rotation"`       // user IDs in rotation order
	RotationDays  int       `json:"rotation_days"`  // how often to rotate
	LastRotation  time.Time `json:"last_rotation"`
}

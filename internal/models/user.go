package models

import "time"

// UserRole represents the user's role within an organization
type UserRole string

const (
	RoleAdmin     UserRole = "admin"
	RoleResponder UserRole = "responder"
	RoleMember    UserRole = "member"
)

// User represents a user account
type User struct {
	ID             string    `json:"id"`
	Email          string    `json:"email"`
	Name           string    `json:"name"`
	PasswordHash   string    `json:"password_hash"`
	Role           UserRole  `json:"role"`
	OrganizationID string    `json:"organization_id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// UserResponse represents user data sent to clients (without password hash)
type UserResponse struct {
	ID             string    `json:"id"`
	Email          string    `json:"email"`
	Name           string    `json:"name"`
	Role           UserRole  `json:"role"`
	OrganizationID string    `json:"organization_id"`
	CreatedAt      time.Time `json:"created_at"`
}

// ToResponse converts User to UserResponse
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:             u.ID,
		Email:          u.Email,
		Name:           u.Name,
		Role:           u.Role,
		OrganizationID: u.OrganizationID,
		CreatedAt:      u.CreatedAt,
	}
}

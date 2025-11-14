package storage

import (
	"fmt"
	"incident-backend/internal/models"
	"path/filepath"
	"strings"
)

// UserStore handles user data persistence
type UserStore struct {
	store *JSONStore
}

// NewUserStore creates a new user store
func NewUserStore(store *JSONStore) *UserStore {
	return &UserStore{store: store}
}

// GetAll retrieves all users for an organization
func (s *UserStore) GetAll(orgID string) ([]models.User, error) {
	filePath := filepath.Join(s.store.GetOrgPath(orgID), "users.json")

	var users []models.User
	if err := s.store.ReadJSON(filePath, &users); err != nil {
		return nil, err
	}

	if users == nil {
		users = []models.User{}
	}

	return users, nil
}

// GetByID retrieves a user by ID
func (s *UserStore) GetByID(orgID, userID string) (*models.User, error) {
	users, err := s.GetAll(orgID)
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		if user.ID == userID {
			return &user, nil
		}
	}

	return nil, fmt.Errorf("user not found")
}

// GetByEmail retrieves a user by email
func (s *UserStore) GetByEmail(orgID, email string) (*models.User, error) {
	users, err := s.GetAll(orgID)
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		if strings.EqualFold(user.Email, email) {
			return &user, nil
		}
	}

	return nil, fmt.Errorf("user not found")
}

// Create creates a new user
func (s *UserStore) Create(user *models.User) error {
	users, err := s.GetAll(user.OrganizationID)
	if err != nil {
		return err
	}

	// Check if email already exists
	for _, existing := range users {
		if strings.EqualFold(existing.Email, user.Email) {
			return fmt.Errorf("user with email '%s' already exists", user.Email)
		}
	}

	users = append(users, *user)

	filePath := filepath.Join(s.store.GetOrgPath(user.OrganizationID), "users.json")
	return s.store.WriteJSON(filePath, users)
}

// Update updates an existing user
func (s *UserStore) Update(user *models.User) error {
	users, err := s.GetAll(user.OrganizationID)
	if err != nil {
		return err
	}

	found := false
	for i, existing := range users {
		if existing.ID == user.ID {
			users[i] = *user
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("user not found")
	}

	filePath := filepath.Join(s.store.GetOrgPath(user.OrganizationID), "users.json")
	return s.store.WriteJSON(filePath, users)
}

// Delete deletes a user
func (s *UserStore) Delete(orgID, userID string) error {
	users, err := s.GetAll(orgID)
	if err != nil {
		return err
	}

	newUsers := []models.User{}
	found := false
	for _, user := range users {
		if user.ID != userID {
			newUsers = append(newUsers, user)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("user not found")
	}

	filePath := filepath.Join(s.store.GetOrgPath(orgID), "users.json")
	return s.store.WriteJSON(filePath, newUsers)
}

// CountByOrg returns the number of users in an organization
func (s *UserStore) CountByOrg(orgID string) (int, error) {
	users, err := s.GetAll(orgID)
	if err != nil {
		return 0, err
	}
	return len(users), nil
}

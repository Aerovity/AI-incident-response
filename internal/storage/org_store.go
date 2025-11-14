package storage

import (
	"fmt"
	"incident-backend/internal/models"
	"path/filepath"
	"strings"
)

// OrgStore handles organization data persistence
type OrgStore struct {
	store *JSONStore
}

// NewOrgStore creates a new organization store
func NewOrgStore(store *JSONStore) *OrgStore {
	return &OrgStore{store: store}
}

// GetAll retrieves all organizations
func (s *OrgStore) GetAll() ([]models.Organization, error) {
	filePath := filepath.Join(s.store.GetGlobalPath(), "organizations.json")

	var orgs []models.Organization
	if err := s.store.ReadJSON(filePath, &orgs); err != nil {
		return nil, err
	}

	if orgs == nil {
		orgs = []models.Organization{}
	}

	return orgs, nil
}

// GetByID retrieves an organization by ID
func (s *OrgStore) GetByID(id string) (*models.Organization, error) {
	orgs, err := s.GetAll()
	if err != nil {
		return nil, err
	}

	for _, org := range orgs {
		if org.ID == id {
			return &org, nil
		}
	}

	return nil, fmt.Errorf("organization not found")
}

// GetBySlug retrieves an organization by slug
func (s *OrgStore) GetBySlug(slug string) (*models.Organization, error) {
	orgs, err := s.GetAll()
	if err != nil {
		return nil, err
	}

	for _, org := range orgs {
		if org.Slug == slug {
			return &org, nil
		}
	}

	return nil, fmt.Errorf("organization not found")
}

// Create creates a new organization
func (s *OrgStore) Create(org *models.Organization) error {
	orgs, err := s.GetAll()
	if err != nil {
		return err
	}

	// Check if slug already exists
	for _, existing := range orgs {
		if strings.EqualFold(existing.Slug, org.Slug) {
			return fmt.Errorf("organization with slug '%s' already exists", org.Slug)
		}
	}

	orgs = append(orgs, *org)

	filePath := filepath.Join(s.store.GetGlobalPath(), "organizations.json")
	if err := s.store.WriteJSON(filePath, orgs); err != nil {
		return err
	}

	// Create organization directory structure
	orgPath := s.store.GetOrgPath(org.ID)
	if err := s.store.ensureDir(orgPath); err != nil {
		return fmt.Errorf("failed to create organization directory: %w", err)
	}

	// Initialize empty data files
	emptyUsers := []models.User{}
	emptyIncidents := []models.Incident{}
	emptyTeams := []models.Team{}
	emptyLearnedFixes := make(map[string]interface{})

	usersPath := filepath.Join(orgPath, "users.json")
	incidentsPath := filepath.Join(orgPath, "incidents.json")
	teamsPath := filepath.Join(orgPath, "teams.json")
	learnedPath := filepath.Join(orgPath, "learned_fixes.json")

	if err := s.store.WriteJSON(usersPath, emptyUsers); err != nil {
		return err
	}
	if err := s.store.WriteJSON(incidentsPath, emptyIncidents); err != nil {
		return err
	}
	if err := s.store.WriteJSON(teamsPath, emptyTeams); err != nil {
		return err
	}
	if err := s.store.WriteJSON(learnedPath, emptyLearnedFixes); err != nil {
		return err
	}

	return nil
}

// Update updates an existing organization
func (s *OrgStore) Update(org *models.Organization) error {
	orgs, err := s.GetAll()
	if err != nil {
		return err
	}

	found := false
	for i, existing := range orgs {
		if existing.ID == org.ID {
			orgs[i] = *org
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("organization not found")
	}

	filePath := filepath.Join(s.store.GetGlobalPath(), "organizations.json")
	return s.store.WriteJSON(filePath, orgs)
}

// Delete deletes an organization
func (s *OrgStore) Delete(id string) error {
	orgs, err := s.GetAll()
	if err != nil {
		return err
	}

	newOrgs := []models.Organization{}
	found := false
	for _, org := range orgs {
		if org.ID != id {
			newOrgs = append(newOrgs, org)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("organization not found")
	}

	filePath := filepath.Join(s.store.GetGlobalPath(), "organizations.json")
	return s.store.WriteJSON(filePath, newOrgs)
}

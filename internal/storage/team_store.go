package storage

import (
	"fmt"
	"incident-backend/internal/models"
	"path/filepath"
)

// TeamStore handles team data persistence
type TeamStore struct {
	store *JSONStore
}

// NewTeamStore creates a new team store
func NewTeamStore(store *JSONStore) *TeamStore {
	return &TeamStore{store: store}
}

// GetAll retrieves all teams for an organization
func (s *TeamStore) GetAll(orgID string) ([]models.Team, error) {
	filePath := filepath.Join(s.store.GetOrgPath(orgID), "teams.json")

	var teams []models.Team
	if err := s.store.ReadJSON(filePath, &teams); err != nil {
		return nil, err
	}

	if teams == nil {
		teams = []models.Team{}
	}

	return teams, nil
}

// GetByID retrieves a team by ID
func (s *TeamStore) GetByID(orgID, teamID string) (*models.Team, error) {
	teams, err := s.GetAll(orgID)
	if err != nil {
		return nil, err
	}

	for _, team := range teams {
		if team.ID == teamID {
			return &team, nil
		}
	}

	return nil, fmt.Errorf("team not found")
}

// Create creates a new team
func (s *TeamStore) Create(team *models.Team) error {
	teams, err := s.GetAll(team.OrganizationID)
	if err != nil {
		return err
	}

	teams = append(teams, *team)

	filePath := filepath.Join(s.store.GetOrgPath(team.OrganizationID), "teams.json")
	return s.store.WriteJSON(filePath, teams)
}

// Update updates an existing team
func (s *TeamStore) Update(team *models.Team) error {
	teams, err := s.GetAll(team.OrganizationID)
	if err != nil {
		return err
	}

	found := false
	for i, existing := range teams {
		if existing.ID == team.ID {
			teams[i] = *team
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("team not found")
	}

	filePath := filepath.Join(s.store.GetOrgPath(team.OrganizationID), "teams.json")
	return s.store.WriteJSON(filePath, teams)
}

// Delete deletes a team
func (s *TeamStore) Delete(orgID, teamID string) error {
	teams, err := s.GetAll(orgID)
	if err != nil {
		return err
	}

	newTeams := []models.Team{}
	found := false
	for _, team := range teams {
		if team.ID != teamID {
			newTeams = append(newTeams, team)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("team not found")
	}

	filePath := filepath.Join(s.store.GetOrgPath(orgID), "teams.json")
	return s.store.WriteJSON(filePath, newTeams)
}

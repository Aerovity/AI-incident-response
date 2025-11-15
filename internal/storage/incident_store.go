package storage

import (
	"fmt"
	"incident-backend/internal/models"
	"path/filepath"
)

// IncidentStore handles incident data persistence
type IncidentStore struct {
	store *JSONStore
}

// NewIncidentStore creates a new incident store
func NewIncidentStore(store *JSONStore) *IncidentStore {
	return &IncidentStore{store: store}
}

// GetAll retrieves all incidents for an organization
func (s *IncidentStore) GetAll(orgID string) ([]models.Incident, error) {
	filePath := filepath.Join(s.store.GetOrgPath(orgID), "incidents.json")

	var incidents []models.Incident
	if err := s.store.ReadJSON(filePath, &incidents); err != nil {
		return nil, err
	}

	if incidents == nil {
		incidents = []models.Incident{}
	}

	return incidents, nil
}

// GetByID retrieves an incident by ID
func (s *IncidentStore) GetByID(orgID, incidentID string) (*models.Incident, error) {
	incidents, err := s.GetAll(orgID)
	if err != nil {
		return nil, err
	}

	for _, incident := range incidents {
		if incident.ID == incidentID {
			return &incident, nil
		}
	}

	return nil, fmt.Errorf("incident not found")
}

// Create creates a new incident
func (s *IncidentStore) Create(incident *models.Incident) error {
	incidents, err := s.GetAll(incident.OrganizationID)
	if err != nil {
		return err
	}

	incidents = append(incidents, *incident)

	filePath := filepath.Join(s.store.GetOrgPath(incident.OrganizationID), "incidents.json")
	return s.store.WriteJSON(filePath, incidents)
}

// Update updates an existing incident
func (s *IncidentStore) Update(incident *models.Incident) error {
	incidents, err := s.GetAll(incident.OrganizationID)
	if err != nil {
		return err
	}

	found := false
	for i, existing := range incidents {
		if existing.ID == incident.ID {
			incidents[i] = *incident
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("incident not found")
	}

	filePath := filepath.Join(s.store.GetOrgPath(incident.OrganizationID), "incidents.json")
	return s.store.WriteJSON(filePath, incidents)
}

// Delete deletes an incident
func (s *IncidentStore) Delete(orgID, incidentID string) error {
	incidents, err := s.GetAll(orgID)
	if err != nil {
		return err
	}

	newIncidents := []models.Incident{}
	found := false
	for _, incident := range incidents {
		if incident.ID != incidentID {
			newIncidents = append(newIncidents, incident)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("incident not found")
	}

	filePath := filepath.Join(s.store.GetOrgPath(orgID), "incidents.json")
	return s.store.WriteJSON(filePath, newIncidents)
}

// Filter retrieves incidents matching the filter criteria
func (s *IncidentStore) Filter(orgID string, status *models.IncidentStatus, priority *models.IncidentPriority, assignedTo *string) ([]models.Incident, error) {
	incidents, err := s.GetAll(orgID)
	if err != nil {
		return nil, err
	}

	filtered := []models.Incident{}
	for _, incident := range incidents {
		// Apply filters
		if status != nil && incident.Status != *status {
			continue
		}
		if priority != nil && incident.Priority != *priority {
			continue
		}
		if assignedTo != nil && (incident.AssignedToID == nil || *incident.AssignedToID != *assignedTo) {
			continue
		}

		filtered = append(filtered, incident)
	}

	return filtered, nil
}

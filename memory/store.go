package memory

import (
	"encoding/json"
	"fmt"
	"incident-ai/models"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

// Store manages incident history and learned fixes
type Store struct {
	incidents map[string]*models.Incident // incident ID -> incident
	fixes     map[string]*models.Resolution // incident type -> successful resolution
	mu        sync.RWMutex
	filePath  string
}

// StoredData represents the data structure saved to disk
type StoredData struct {
	Incidents   map[string]*models.Incident   `json:"incidents"`
	Fixes       map[string]*models.Resolution `json:"fixes"`
	LastUpdated time.Time                     `json:"last_updated"`
}

// NewStore creates a new memory store
func NewStore(filePath string) *Store {
	store := &Store{
		incidents: make(map[string]*models.Incident),
		fixes:     make(map[string]*models.Resolution),
		filePath:  filePath,
	}

	// Try to load existing data
	if err := store.Load(); err != nil {
		log.Printf("[MEMORY] No existing data found, starting fresh: %v\n", err)
	} else {
		log.Printf("[MEMORY] Loaded %d incidents and %d learned fixes\n",
			len(store.incidents), len(store.fixes))
	}

	return store
}

// StoreIncident saves an incident to memory
func (s *Store) StoreIncident(incident *models.Incident) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.incidents[incident.ID] = incident

	// If incident was resolved successfully, store the fix for future use
	if incident.Status == models.StatusResolved && incident.Resolution != nil && incident.Resolution.Success {
		s.fixes[string(incident.Type)] = incident.Resolution
		log.Printf("[MEMORY] Learned fix for %s incidents\n", incident.Type)
	}

	return s.save()
}

// GetIncident retrieves an incident by ID
func (s *Store) GetIncident(id string) (*models.Incident, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	incident, exists := s.incidents[id]
	if !exists {
		return nil, fmt.Errorf("incident not found: %s", id)
	}

	return incident, nil
}

// GetLearnedFix checks if we have a learned fix for this incident type
func (s *Store) GetLearnedFix(incidentType models.IncidentType) (*models.Resolution, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	fix, exists := s.fixes[string(incidentType)]
	return fix, exists
}

// HasLearnedFix checks if we have a fix for this incident type
func (s *Store) HasLearnedFix(incidentType models.IncidentType) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.fixes[string(incidentType)]
	return exists
}

// GetAllIncidents returns all stored incidents
func (s *Store) GetAllIncidents() []*models.Incident {
	s.mu.RLock()
	defer s.mu.RUnlock()

	incidents := make([]*models.Incident, 0, len(s.incidents))
	for _, incident := range s.incidents {
		incidents = append(incidents, incident)
	}

	return incidents
}

// GetStats returns statistics about stored incidents
func (s *Store) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	totalIncidents := len(s.incidents)
	resolvedCount := 0
	failedCount := 0
	typeCount := make(map[string]int)

	for _, incident := range s.incidents {
		typeCount[string(incident.Type)]++

		if incident.Status == models.StatusResolved {
			resolvedCount++
		} else if incident.Status == models.StatusFailed {
			failedCount++
		}
	}

	return map[string]interface{}{
		"total_incidents":    totalIncidents,
		"resolved":           resolvedCount,
		"failed":             failedCount,
		"learned_fixes":      len(s.fixes),
		"incidents_by_type":  typeCount,
		"available_fix_types": s.getFixTypes(),
	}
}

func (s *Store) getFixTypes() []string {
	types := make([]string, 0, len(s.fixes))
	for t := range s.fixes {
		types = append(types, t)
	}
	return types
}

// Save persists the store to disk
func (s *Store) save() error {
	data := StoredData{
		Incidents:   s.incidents,
		Fixes:       s.fixes,
		LastUpdated: time.Now(),
	}

	file, err := os.Create(s.filePath)
	if err != nil {
		return fmt.Errorf("failed to create store file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode store data: %w", err)
	}

	return nil
}

// Load reads the store from disk
func (s *Store) Load() error {
	file, err := os.Open(s.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	var data StoredData
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return fmt.Errorf("failed to decode store data: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.incidents = data.Incidents
	s.fixes = data.Fixes

	return nil
}

// Clear removes all data from the store
func (s *Store) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.incidents = make(map[string]*models.Incident)
	s.fixes = make(map[string]*models.Resolution)

	return s.save()
}

// UpdateIncidentStatus updates the status of an incident
func (s *Store) UpdateIncidentStatus(id string, status models.IncidentStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	incident, exists := s.incidents[id]
	if !exists {
		return fmt.Errorf("incident not found: %s", id)
	}

	incident.Status = status

	if status == models.StatusResolved {
		now := time.Now()
		incident.ResolvedAt = &now
	}

	return s.save()
}

// PrintSummary prints a summary of stored incidents
func (s *Store) PrintSummary() {
	stats := s.GetStats()

	log.Println("\n" + strings.Repeat("=", 70))
	log.Println("[MEMORY] Incident Response System - Summary")
	log.Println(strings.Repeat("=", 70))
	log.Printf("Total Incidents Handled: %v\n", stats["total_incidents"])
	log.Printf("Successfully Resolved:   %v\n", stats["resolved"])
	log.Printf("Failed:                  %v\n", stats["failed"])
	log.Printf("Learned Fixes Available: %v\n", stats["learned_fixes"])

	if fixTypes, ok := stats["available_fix_types"].([]string); ok && len(fixTypes) > 0 {
		log.Println("\nLearned fixes for incident types:")
		for _, t := range fixTypes {
			log.Printf("  ✓ %s\n", t)
		}
	}

	log.Println(strings.Repeat("=", 70) + "\n")
}

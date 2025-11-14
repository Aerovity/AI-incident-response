package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// JSONStore handles generic JSON file operations with file locking
type JSONStore struct {
	dataDir string
	mu      sync.RWMutex
}

// NewJSONStore creates a new JSON store
func NewJSONStore(dataDir string) *JSONStore {
	return &JSONStore{
		dataDir: dataDir,
	}
}

// ensureDir creates a directory if it doesn't exist
func (s *JSONStore) ensureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// ReadJSON reads and unmarshals a JSON file
func (s *JSONStore) ReadJSON(filePath string, v interface{}) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist yet, not an error
		}
		return fmt.Errorf("failed to read file: %w", err)
	}

	if len(data) == 0 {
		return nil // Empty file, not an error
	}

	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return nil
}

// WriteJSON marshals and writes data to a JSON file
func (s *JSONStore) WriteJSON(filePath string, v interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := s.ensureDir(dir); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Write to a temporary file first, then rename (atomic operation)
	tmpFile := filePath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := os.Rename(tmpFile, filePath); err != nil {
		os.Remove(tmpFile) // Clean up temp file on error
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// GetOrgPath returns the path to an organization's data directory
func (s *JSONStore) GetOrgPath(orgID string) string {
	return filepath.Join(s.dataDir, "organizations", orgID)
}

// GetGlobalPath returns the path to global data files
func (s *JSONStore) GetGlobalPath() string {
	return filepath.Join(s.dataDir, "global")
}

// FileExists checks if a file exists
func (s *JSONStore) FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

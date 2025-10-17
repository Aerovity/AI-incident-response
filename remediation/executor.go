package remediation

import (
	"fmt"
	"incident-ai/models"
	"incident-ai/service"
	"log"
	"strings"
	"time"
)

// Executor applies fixes to resolve incidents
type Executor struct {
	targetService *service.TargetService
}

// NewExecutor creates a new remediation executor
func NewExecutor(targetService *service.TargetService) *Executor {
	return &Executor{
		targetService: targetService,
	}
}

// ExecuteFix applies the AI-suggested fix
func (e *Executor) ExecuteFix(incident *models.Incident, aiResponse *models.AIResponse) (*models.Resolution, error) {
	log.Printf("[REMEDIATION] Applying fix for incident %s (Type: %s)\n", incident.ID, aiResponse.FixType)

	resolution := &models.Resolution{
		FixType:     aiResponse.FixType,
		Description: aiResponse.Diagnosis,
		Steps:       aiResponse.FixSteps,
		Code:        aiResponse.Code,
		Success:     false,
	}

	var err error

	switch aiResponse.FixType {
	case "restart":
		err = e.executeRestart(aiResponse.FixSteps)
	case "config":
		err = e.executeConfigFix(aiResponse.FixSteps)
	case "code":
		err = e.executeCodeFix(aiResponse)
	default:
		err = fmt.Errorf("unknown fix type: %s", aiResponse.FixType)
	}

	if err != nil {
		log.Printf("[REMEDIATION] ❌ Fix failed: %v\n", err)
		resolution.Success = false
		return resolution, err
	}

	resolution.Success = true
	log.Println("[REMEDIATION] ✓ Fix applied successfully")

	return resolution, nil
}

func (e *Executor) executeRestart(steps []string) error {
	log.Println("[REMEDIATION] Executing restart fix...")

	for i, step := range steps {
		log.Printf("[REMEDIATION]   Step %d: %s\n", i+1, step)
	}

	// Stop the service
	if e.targetService.IsHealthy() || true { // Always try to stop
		log.Println("[REMEDIATION]   → Stopping service...")
		if err := e.targetService.Stop(); err != nil {
			log.Printf("[REMEDIATION]   → Stop error (continuing): %v\n", err)
		}
		time.Sleep(500 * time.Millisecond)
	}

	// Start the service
	log.Println("[REMEDIATION]   → Starting service...")
	if err := e.targetService.Start(); err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}

	time.Sleep(1 * time.Second) // Give service time to fully start

	log.Println("[REMEDIATION]   → Service restarted")
	return nil
}

func (e *Executor) executeConfigFix(steps []string) error {
	log.Println("[REMEDIATION] Executing config fix...")

	for i, step := range steps {
		log.Printf("[REMEDIATION]   Step %d: %s\n", i+1, step)

		// Parse the step to extract config changes
		if err := e.applyConfigStep(step); err != nil {
			log.Printf("[REMEDIATION]   → Error: %v\n", err)
		}
	}

	// Always restart after config changes
	log.Println("[REMEDIATION]   → Restarting service to apply config changes...")
	return e.targetService.Restart()
}

func (e *Executor) applyConfigStep(step string) error {
	step = strings.ToLower(step)

	// Look for common config patterns in the step description
	if strings.Contains(step, "database_url") || strings.Contains(step, "database url") {
		if strings.Contains(step, "localhost:5432") || strings.Contains(step, "restore") {
			log.Println("[REMEDIATION]     → Restoring database_url to localhost:5432")
			e.targetService.SetConfig("database_url", "localhost:5432")
			return nil
		}
	}

	if strings.Contains(step, "timeout") {
		if strings.Contains(step, "30s") || strings.Contains(step, "restore") || strings.Contains(step, "reset") {
			log.Println("[REMEDIATION]     → Restoring timeout to 30s")
			e.targetService.SetConfig("timeout", "30s")
			return nil
		}
	}

	if strings.Contains(step, "max_retries") || strings.Contains(step, "retries") {
		if strings.Contains(step, "3") || strings.Contains(step, "restore") {
			log.Println("[REMEDIATION]     → Restoring max_retries to 3")
			e.targetService.SetConfig("max_retries", "3")
			return nil
		}
	}

	// If it's a restart step, skip it (will be done after all config changes)
	if strings.Contains(step, "restart") {
		return nil
	}

	// If we can't parse the step, log it but don't error
	log.Printf("[REMEDIATION]     → Config step noted: %s\n", step)
	return nil
}

func (e *Executor) executeCodeFix(aiResponse *models.AIResponse) error {
	log.Println("[REMEDIATION] Executing code fix...")
	log.Println("[REMEDIATION]   ⚠️  Code fixes require manual intervention")
	log.Println("[REMEDIATION]   Code provided by AI:")
	log.Println("[REMEDIATION]   " + strings.Repeat("-", 60))

	if aiResponse.Code != "" {
		// Print code with indentation
		codeLines := strings.Split(aiResponse.Code, "\n")
		for _, line := range codeLines {
			log.Printf("[REMEDIATION]   %s\n", line)
		}
	} else {
		log.Println("[REMEDIATION]   (No code provided)")
	}

	log.Println("[REMEDIATION]   " + strings.Repeat("-", 60))

	// For demo purposes, we'll apply a generic fix
	log.Println("[REMEDIATION]   → Attempting restart as fallback...")
	return e.targetService.Restart()
}

// ApplyCachedFix applies a previously successful fix
func (e *Executor) ApplyCachedFix(incident *models.Incident, cachedResolution *models.Resolution) error {
	log.Printf("[REMEDIATION] Applying cached fix for incident %s\n", incident.ID)
	log.Println("[REMEDIATION] ⚡ Using learned solution (no AI call needed)")

	var err error

	switch cachedResolution.FixType {
	case "restart":
		err = e.executeRestart(cachedResolution.Steps)
	case "config":
		err = e.executeConfigFix(cachedResolution.Steps)
	case "code":
		log.Println("[REMEDIATION] ⚠️  Code fixes cannot be auto-applied from cache")
		err = e.targetService.Restart()
	default:
		err = fmt.Errorf("unknown fix type: %s", cachedResolution.FixType)
	}

	if err != nil {
		log.Printf("[REMEDIATION] ❌ Cached fix failed: %v\n", err)
		return err
	}

	log.Println("[REMEDIATION] ✓ Cached fix applied successfully")
	return nil
}

// GetStatus returns current status of the service
func (e *Executor) GetStatus() map[string]interface{} {
	return map[string]interface{}{
		"service_healthy": e.targetService.IsHealthy(),
		"configuration":   e.targetService.GetConfig(),
		"recent_logs":     e.targetService.GetLogs(),
	}
}

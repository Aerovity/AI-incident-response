package main

import (
	"context"
	"flag"
	"fmt"
	"incident-ai/ai"
	"incident-ai/memory"
	"incident-ai/models"
	"incident-ai/monitor"
	"incident-ai/remediation"
	"incident-ai/service"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

const (
	servicePort    = "8080"
	checkInterval  = 3 * time.Second
	memoryFile     = "incident_memory.json"
)

func main() {
	// Load .env file if it exists
	_ = godotenv.Load()

	// Command line flags
	apiKey := flag.String("api-key", os.Getenv("OPENAI_API_KEY"), "OpenAI API key (or set OPENAI_API_KEY env var)")
	demo := flag.Bool("demo", false, "Run automated demo scenario")
	useAI := flag.Bool("use-ai", true, "Use OpenAI for analysis (false = use fallback logic)")
	flag.Parse()

	printBanner()

	// Validate API key if AI is enabled
	if *useAI && *apiKey == "" {
		log.Println("⚠️  No OpenAI API key provided. Using fallback analysis mode.")
		log.Println("   To use OpenAI: set OPENAI_API_KEY env var or use -api-key flag")
		*useAI = false
	}

	// Initialize components
	log.Println("\n[SYSTEM] Initializing Incident Response System...")

	targetService := service.NewTargetService(servicePort)
	analyzer := ai.NewAnalyzer(*apiKey)
	executor := remediation.NewExecutor(targetService)
	store := memory.NewStore(memoryFile)
	detector := monitor.NewIncidentDetector(
		fmt.Sprintf("http://localhost:%s", servicePort),
		checkInterval,
	)

	// Start target service
	log.Println("[SYSTEM] Starting target service...")
	if err := targetService.Start(); err != nil {
		log.Fatalf("Failed to start service: %v", err)
	}

	// Create orchestrator
	orch := &Orchestrator{
		service:  targetService,
		detector: detector,
		analyzer: analyzer,
		executor: executor,
		store:    store,
		useAI:    *useAI,
	}

	// Setup context and signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start monitoring
	detector.Start(ctx)

	// Start incident handler
	go orch.handleIncidents(ctx)

	log.Println("[SYSTEM] ✓ System ready!")
	log.Printf("[SYSTEM] Service running at: http://localhost:%s\n", servicePort)
	log.Println("\n" + strings.Repeat("=", 70))
	printUsageInstructions()

	// Run demo if requested
	if *demo {
		go runDemo(targetService)
	}

	// Wait for interrupt
	<-sigChan
	log.Println("\n[SYSTEM] Shutting down...")

	cancel()
	detector.Stop()
	targetService.Stop()

	log.Println("[SYSTEM] Printing final summary...")
	store.PrintSummary()

	log.Println("[SYSTEM] Goodbye!")
}

// Orchestrator coordinates incident detection and response
type Orchestrator struct {
	service  *service.TargetService
	detector *monitor.IncidentDetector
	analyzer *ai.Analyzer
	executor *remediation.Executor
	store    *memory.Store
	useAI    bool
}

func (o *Orchestrator) handleIncidents(ctx context.Context) {
	incidentChan := o.detector.GetIncidentChannel()

	for {
		select {
		case <-ctx.Done():
			return

		case incident := <-incidentChan:
			if err := o.processIncident(ctx, incident); err != nil {
				log.Printf("[SYSTEM] ❌ Failed to process incident: %v\n", err)
			}
		}
	}
}

func (o *Orchestrator) processIncident(ctx context.Context, incident *models.Incident) error {
	log.Println("\n" + strings.Repeat("=", 70))
	log.Printf("[DETECTOR] 🚨 Incident Detected: %s\n", incident.Type)
	log.Printf("[DETECTOR] ID: %s\n", incident.ID)
	log.Println(strings.Repeat("=", 70))

	// Store initial incident
	if err := o.store.StoreIncident(incident); err != nil {
		log.Printf("[MEMORY] Warning: failed to store incident: %v\n", err)
	}

	// Check if we have a learned fix
	if cachedFix, exists := o.store.GetLearnedFix(incident.Type); exists {
		log.Println("[MEMORY] ⚡ Found learned fix! Applying without AI call...")
		incident.UsedCachedFix = true

		if err := o.executor.ApplyCachedFix(incident, cachedFix); err != nil {
			log.Printf("[REMEDIATION] ❌ Cached fix failed: %v\n", err)
			log.Println("[REMEDIATION] Falling back to AI analysis...")
		} else {
			// Verify resolution
			if o.verifyResolution() {
				incident.Status = models.StatusResolved
				now := time.Now()
				incident.ResolvedAt = &now
				incident.Resolution = cachedFix
				o.store.StoreIncident(incident)

				log.Println("[SYSTEM] ✅ Incident resolved using cached fix!")
				log.Printf("[SYSTEM] Resolution time: %v\n", time.Since(incident.DetectedAt))
				return nil
			} else {
				log.Println("[VERIFICATION] ❌ Service still unhealthy after cached fix")
			}
		}
	}

	// No cached fix or cached fix failed - use AI
	incident.Status = models.StatusAnalyzing
	o.store.UpdateIncidentStatus(incident.ID, models.StatusAnalyzing)

	var aiResponse *models.AIResponse
	var err error

	if o.useAI {
		log.Println("[AI] Calling OpenAI for incident analysis...")
		aiResponse, err = o.analyzer.AnalyzeIncident(ctx, incident)
		if err != nil {
			log.Printf("[AI] ❌ OpenAI error: %v\n", err)
			log.Println("[AI] Falling back to rule-based analysis...")
			aiResponse = o.analyzer.GetQuickAnalysis(incident)
		}
	} else {
		log.Println("[AI] Using fallback rule-based analysis...")
		aiResponse = o.analyzer.GetQuickAnalysis(incident)
	}

	incident.Diagnosis = aiResponse.Diagnosis
	log.Printf("[AI] 📊 Diagnosis: %s\n", aiResponse.Diagnosis)
	log.Printf("[AI] 🔧 Fix Type: %s\n", aiResponse.FixType)
	log.Printf("[AI] 📝 Steps: %d\n", len(aiResponse.FixSteps))

	// Execute fix
	incident.Status = models.StatusFixing
	o.store.UpdateIncidentStatus(incident.ID, models.StatusFixing)

	resolution, err := o.executor.ExecuteFix(incident, aiResponse)
	if err != nil {
		incident.Status = models.StatusFailed
		o.store.StoreIncident(incident)
		return fmt.Errorf("failed to execute fix: %w", err)
	}

	incident.Resolution = resolution

	// Verify resolution
	time.Sleep(2 * time.Second) // Give service time to stabilize

	if o.verifyResolution() {
		incident.Status = models.StatusResolved
		now := time.Now()
		incident.ResolvedAt = &now
		o.store.StoreIncident(incident)

		log.Println("\n" + strings.Repeat("=", 70))
		log.Println("[SYSTEM] ✅ INCIDENT RESOLVED!")
		log.Printf("[SYSTEM] Resolution time: %v\n", time.Since(incident.DetectedAt))
		log.Println(strings.Repeat("=", 70) + "\n")
	} else {
		incident.Status = models.StatusFailed
		o.store.StoreIncident(incident)

		log.Println("\n" + strings.Repeat("=", 70))
		log.Println("[SYSTEM] ❌ INCIDENT NOT RESOLVED")
		log.Println("[SYSTEM] Service still reporting unhealthy after fix attempt")
		log.Println(strings.Repeat("=", 70) + "\n")
	}

	return nil
}

func (o *Orchestrator) verifyResolution() bool {
	log.Println("[VERIFICATION] Checking service health...")

	// Multiple checks to ensure stability
	for i := 0; i < 3; i++ {
		if i > 0 {
			time.Sleep(1 * time.Second)
		}

		if o.detector.VerifyResolution() {
			log.Printf("[VERIFICATION] ✓ Health check %d/3 passed\n", i+1)
		} else {
			log.Printf("[VERIFICATION] ✗ Health check %d/3 failed\n", i+1)
			return false
		}
	}

	log.Println("[VERIFICATION] ✅ All health checks passed!")
	return true
}

func printBanner() {
	banner := `
╔═══════════════════════════════════════════════════════════════════╗
║                                                                   ║
║        🤖 AI-Powered Incident Response System                    ║
║                                                                   ║
║        Automatic Detection • AI Analysis • Smart Remediation     ║
║                                                                   ║
╚═══════════════════════════════════════════════════════════════════╝
`
	fmt.Println(banner)
}

func printUsageInstructions() {
	instructions := `
📋 HOW TO USE:

1. Trigger an incident:
   curl "http://localhost:8080/trigger-incident?type=crash"

   Available incident types:
   • crash      - Service crashes/stops responding
   • config     - Configuration becomes corrupted
   • resource   - Resource exhaustion (port/memory)
   • dependency - External dependency failure

2. Watch the system:
   • Automatically detect the incident
   • Analyze with AI (or use learned fix)
   • Apply remediation
   • Verify resolution

3. Trigger the same incident again to see it use the cached fix!

4. Check service status:
   curl http://localhost:8080/status

5. Press Ctrl+C to stop and see summary

` + strings.Repeat("=", 70) + "\n"

	fmt.Println(instructions)
}

func runDemo(targetService *service.TargetService) {
	log.Println("\n[DEMO] Starting automated demo in 5 seconds...")
	time.Sleep(5 * time.Second)

	incidents := []struct {
		name     string
		typeStr  string
		waitTime time.Duration
	}{
		{"Service Crash", "crash", 15 * time.Second},
		{"Config Error", "config", 15 * time.Second},
		{"Service Crash (cached)", "crash", 15 * time.Second},
		{"Dependency Failure", "dependency", 15 * time.Second},
	}

	for i, inc := range incidents {
		log.Printf("\n[DEMO] (%d/%d) Triggering: %s\n", i+1, len(incidents), inc.name)

		// Trigger incident via internal API
		targetService.Stop()
		time.Sleep(500 * time.Millisecond)
		targetService.Start()
		time.Sleep(1 * time.Second)

		// Trigger the incident
		client := &http.Client{}
		url := fmt.Sprintf("http://localhost:%s/trigger-incident?type=%s", servicePort, inc.typeStr)
		resp, err := client.Get(url)
		if err != nil {
			log.Printf("[DEMO] Failed to trigger incident: %v\n", err)
		} else {
			resp.Body.Close()
		}

		// Wait for resolution
		log.Printf("[DEMO] Waiting %v for resolution...\n", inc.waitTime)
		time.Sleep(inc.waitTime)
	}

	log.Println("\n[DEMO] Demo complete! Press Ctrl+C to see summary.")
}

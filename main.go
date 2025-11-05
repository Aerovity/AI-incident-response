package main

import (
	"context"
	"flag"
	"fmt"
	"incident-ai/ai"
	"incident-ai/analytics"
	"incident-ai/cloudflare"
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
	demo := flag.Bool("demo", false, "Run automated demo scenario")
	useAI := flag.Bool("use-ai", true, "Use Cloudflare AI for analysis (false = use fallback logic)")
	setupWrangler := flag.Bool("setup", false, "Setup wrangler configuration")
	flag.Parse()

	printBanner()

	// Get Cloudflare credentials via wrangler
	var apiKey, accountID string
	if *useAI {
		config, err := cloudflare.GetCredentials()
		if err != nil {
			log.Println("\n⚠️  Cloudflare Authentication Not Found")
			log.Println("=" + strings.Repeat("=", 60))
			log.Println("\nTo use Cloudflare AI, you need to authenticate:")
			log.Println("\n1. Install wrangler (if not already installed):")
			log.Println("   npm install -g wrangler")
			log.Println("\n2. Login to Cloudflare:")
			log.Println("   npx wrangler login")
			log.Println("\n3. (Optional) Add your account ID to wrangler.toml:")
			log.Println("   Run: npx wrangler whoami")
			log.Println("   Copy your Account ID and add to wrangler.toml")
			log.Println("\nAlternatively, set environment variables:")
			log.Println("   CLOUDFLARE_API_KEY=your-key")
			log.Println("   CLOUDFLARE_ACCOUNT_ID=your-account-id")
			log.Println("\nOr run without AI:")
			log.Println("   go run main.go -use-ai=false")
			log.Println("\n" + strings.Repeat("=", 60) + "\n")

			*useAI = false
			apiKey = ""
			accountID = ""
		} else {
			apiKey = config.APIKey
			accountID = config.AccountID
			log.Println("✓ Cloudflare credentials loaded successfully")
		}
	}

	// Setup mode
	if *setupWrangler {
		if err := setupWranglerAuth(); err != nil {
			log.Fatalf("Setup failed: %v", err)
		}
		return
	}

	// Initialize components
	log.Println("\n[SYSTEM] Initializing Incident Response System...")

	targetService := service.NewTargetService(servicePort)
	analyzer := ai.NewAnalyzer(apiKey, accountID)
	executor := remediation.NewExecutor(targetService)
	store := memory.NewStore(memoryFile)
	analyticsEngine := analytics.NewEngine()
	detector := monitor.NewIncidentDetector(
		fmt.Sprintf("http://localhost:%s", servicePort),
		checkInterval,
	)

	// Start target service
	log.Println("[SYSTEM] Starting target service...")
	if err := targetService.Start(); err != nil {
		log.Fatalf("Failed to start service: %v", err)
	}

	// Set up metrics endpoint
	targetService.SetMetricsFunc(func() map[string]interface{} {
		stats := store.GetStats()
		stats["service_uptime"] = time.Since(time.Now().Add(-10 * time.Second)).String() // Simplified
		return stats
	})

	// Set up analytics endpoint
	targetService.SetAnalyticsFunc(func() interface{} {
		incidents := store.GetAllIncidents()
		report := analyticsEngine.GenerateReport(incidents)
		return report
	})

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
	log.Printf("[DETECTOR] Priority: %s\n", incident.Priority)
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
		log.Println("[AI] Calling Cloudflare AI for incident analysis...")
		aiResponse, err = o.analyzer.AnalyzeIncident(ctx, incident)
		if err != nil {
			log.Printf("[AI] ❌ Cloudflare AI error: %v\n", err)
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
║        🔥 AI-Powered Incident Response System                    ║
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
   • crash      - Service crashes/stops responding (CRITICAL)
   • config     - Configuration becomes corrupted (MEDIUM)
   • resource   - Resource exhaustion (port/memory) (MEDIUM)
   • dependency - External dependency failure (HIGH)
   • database   - Database errors and timeouts (CRITICAL)
   • latency    - High response latency (LOW)
   • memory     - Memory leak detected (MEDIUM)
   • security   - Security breach attempt (CRITICAL)
   • network    - Network partition (HIGH)
   • disk       - Disk storage full (HIGH)

2. Watch the system:
   • Automatically detect the incident
   • Analyze with AI (or use learned fix)
   • Apply remediation with priority-based handling
   • Verify resolution

3. Trigger the same incident again to see it use the cached fix!

4. Check service status:
   curl http://localhost:8080/status

5. View basic metrics:
   curl http://localhost:8080/metrics

6. View advanced analytics & trends:
   curl http://localhost:8080/analytics

7. Press Ctrl+C to stop and see summary

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

func setupWranglerAuth() error {
	log.Println("\n🔧 Wrangler Setup Wizard")
	log.Println("=" + strings.Repeat("=", 60))

	// Check if wrangler is installed
	log.Println("\n[1/3] Checking for wrangler...")
	if !cloudflare.CheckWranglerInstalled() {
		log.Println("❌ Wrangler not found. Please install it first:")
		log.Println("   npm install -g wrangler")
		return fmt.Errorf("wrangler not installed")
	}
	log.Println("✓ Wrangler is installed")

	// Get account ID from wrangler
	log.Println("\n[2/3] Getting your Cloudflare Account ID...")
	accountID, err := cloudflare.GetAccountIDFromWrangler()
	if err != nil {
		log.Println("⚠️  Could not get account ID automatically.")
		log.Println("   Please run: npx wrangler whoami")
		log.Println("   Then manually add account_id to wrangler.toml")
		return err
	}
	log.Printf("✓ Found Account ID: %s\n", accountID)

	// Update wrangler.toml
	log.Println("\n[3/3] Updating wrangler.toml...")
	if err := cloudflare.SetupWranglerConfig(accountID); err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}
	log.Println("✓ wrangler.toml updated successfully")

	log.Println("\n" + strings.Repeat("=", 60))
	log.Println("✅ Setup complete! You can now run:")
	log.Println("   go run main.go")
	log.Println("=" + strings.Repeat("=", 60) + "\n")

	return nil
}

# 🔥 AI-Powered Incident Response System

An intelligent incident detection and auto-remediation system built in Go, inspired by incident.io's auto-remediation capabilities. The system uses Cloudflare AI (Llama 3.3) to analyze incidents and automatically apply fixes, while learning from past incidents to respond faster in the future.

## ✨ Features

- **🔍 Automatic Incident Detection**: Continuous health monitoring and incident detection
- **🤖 AI-Powered Analysis**: Uses Cloudflare AI (Llama 3.3) to diagnose root causes and suggest fixes
- **⚡ Smart Remediation**: Automatically applies fixes to resolve incidents
- **🧠 Learning System**: Remembers successful fixes and applies them instantly on recurrence
- **📊 Extended Incident Types**: Handles 10+ incident types including crashes, config errors, database issues, security breaches, and more
- **🎯 Incident Prioritization**: Automatic priority assignment (Critical, High, Medium, Low)
- **⏮️ Rollback Capabilities**: Capture state before fixes and rollback if needed
- **🐳 Docker Integration**: Full Docker and docker-compose support
- **📈 Metrics & Analytics**: HTTP endpoint for real-time statistics and insights
- **🔧 Custom Remediation Scripts**: Support for user-defined bash/batch scripts per incident type
- **✅ Verification**: Confirms incidents are truly resolved before marking as complete
- **💾 Persistent Memory**: Stores incident history and learned fixes to disk

## 🏗️ Architecture

```
┌─────────────┐      ┌──────────────┐      ┌─────────────┐
│   Target    │◄─────┤   Monitor/   │─────►│     AI      │
│   Service   │      │   Detector   │      │  Analyzer   │
└─────────────┘      └──────────────┘      └─────────────┘
       ▲                     │                      │
       │                     ▼                      ▼
       │             ┌──────────────┐      ┌─────────────┐
       └─────────────┤ Remediation  │◄─────┤   Memory    │
                     │  Executor    │      │    Store    │
                     └──────────────┘      └─────────────┘
```

### Components

1. **Target Service** (`service/`) - Simulated HTTP service that can experience incidents
2. **Monitor/Detector** (`monitor/`) - Polls service health and detects incidents
3. **AI Analyzer** (`ai/`) - Integrates with Cloudflare AI to analyze incidents and suggest fixes
4. **Remediation Executor** (`remediation/`) - Applies fixes to resolve incidents
5. **Memory Store** (`memory/`) - Stores incident history and learned fixes
6. **Models** (`models/`) - Core data structures

## 🚀 Quick Start

### Prerequisites

- Go 1.21 or higher
- Cloudflare Account with API access (optional - system works with fallback logic if not provided)
- Cloudflare API Key and Account ID

### Installation

1. Clone or download this project:

2. Install dependencies:
```bash
go mod download
```

3. Set your Cloudflare credentials (optional):
```bash
# Windows
set CLOUDFLARE_API_KEY=your-api-key-here
set CLOUDFLARE_ACCOUNT_ID=your-account-id-here

# Linux/Mac
export CLOUDFLARE_API_KEY=your-api-key-here
export CLOUDFLARE_ACCOUNT_ID=your-account-id-here
```

**How to get Cloudflare credentials:**
1. Log in to your Cloudflare dashboard
2. Go to "Workers & Pages" → "Overview"
3. Your Account ID is displayed on the right side
4. For API Key: Go to "My Profile" → "API Tokens" → "Create Token"
5. Use the "Workers AI" template or create a custom token with Workers AI permissions

### Running the System

**Basic mode (with Cloudflare AI):**
```bash
go run main.go
```

**Without Cloudflare AI (fallback mode):**
```bash
go run main.go -use-ai=false
```

**Automated demo:**
```bash
go run main.go -demo
```

**With explicit credentials:**
```bash
go run main.go -api-key=your-api-key -account-id=your-account-id
```

## 📖 Usage

### 1. Trigger an Incident

Once the system is running, trigger incidents using curl:

```bash
# Service crash
curl "http://localhost:8080/trigger-incident?type=crash"

# Configuration error
curl "http://localhost:8080/trigger-incident?type=config"

# Resource exhaustion
curl "http://localhost:8080/trigger-incident?type=resource"

# Dependency failure
curl "http://localhost:8080/trigger-incident?type=dependency"
```

### 2. Watch the Magic

The system will:
1. 🔍 Detect the unhealthy service
2. 📋 Check if it has seen this incident type before
3. 🤖 If new: Ask Cloudflare AI for diagnosis and fix
4. ⚡ If known: Apply cached fix instantly (no AI call needed!)
5. 🔧 Execute the remediation steps
6. ✅ Verify the service is healthy again
7. 💾 Store the successful fix for future use

### 3. Test the Learning System

Trigger the same incident type twice:

```bash
# First time - uses AI
curl "http://localhost:8080/trigger-incident?type=crash"
# Wait for resolution...
#after cash way faster 
curl "http://localhost:8080/trigger-incident?type=crash"
```

The second incident will be resolved instantly using the learned fix!

### 4. Check Service Status

```bash
curl http://localhost:8080/status
```

### 5. View Advanced Analytics

```bash
curl http://localhost:8080/analytics
```

Get comprehensive analytics including:
- Incident trends (hourly/daily)
- Resolution time statistics
- Success rates and predictions
- Hot spots and problem areas

### 6. View Summary

Press `Ctrl+C` to stop the system and see a summary of all incidents handled.

## 🎯 Example Output

```
[MONITOR] ⚠️  Health check FAILED - Incident detected!
══════════════════════════════════════════════════════════════════════
[DETECTOR] 🚨 Incident Detected: SERVICE_DOWN
[DETECTOR] ID: 550e8400-e29b-41d4-a716-446655440000
══════════════════════════════════════════════════════════════════════
[MEMORY] No cached fix found - using AI analysis
[AI] Calling Cloudflare AI for incident analysis...
[AI] 📊 Diagnosis: Service process has crashed or stopped responding
[AI] 🔧 Fix Type: restart
[AI] 📝 Steps: 3
[REMEDIATION] Applying fix for incident 550e8400-e29b-41d4-a716-446655440000
[REMEDIATION]   Step 1: Stop the service if it's still partially running
[REMEDIATION]   Step 2: Restart the service process
[REMEDIATION]   Step 3: Verify health check passes
[REMEDIATION]   → Stopping service...
[REMEDIATION]   → Starting service...
[REMEDIATION]   → Service restarted
[REMEDIATION] ✓ Fix applied successfully
[VERIFICATION] Checking service health...
[VERIFICATION] ✓ Health check 1/3 passed
[VERIFICATION] ✓ Health check 2/3 passed
[VERIFICATION] ✓ Health check 3/3 passed
[VERIFICATION] ✅ All health checks passed!
══════════════════════════════════════════════════════════════════════
[SYSTEM] ✅ INCIDENT RESOLVED!
[SYSTEM] Resolution time: 8.234s
══════════════════════════════════════════════════════════════════════
[MEMORY] Learned fix for SERVICE_DOWN incidents
```

## 🔥 Incident Types

### Critical Priority
1. **Service Crash (`crash` / `SERVICE_DOWN`)**
   - Service stops responding to health checks
   - Typical Fix: Restart the service
   - Priority: **CRITICAL**

2. **Security Breach (`security` / `SECURITY_BREACH`)**
   - Unauthorized access attempts detected
   - Typical Fix: Block access, restart service, audit logs
   - Priority: **CRITICAL**

3. **Database Error (`database` / `DATABASE_ERROR`)**
   - Query timeouts, connection pool exhausted
   - Typical Fix: Reset connections, restart service
   - Priority: **CRITICAL**

### High Priority
4. **Dependency Failure (`dependency` / `DEPENDENCY_FAILURE`)**
   - External dependency unreachable
   - Typical Fix: Fix connection string and reconnect
   - Priority: **HIGH**

5. **Network Partition (`network` / `NETWORK_PARTITION`)**
   - Cluster nodes unreachable
   - Typical Fix: Restore network connectivity
   - Priority: **HIGH**

6. **Disk Full (`disk` / `DISK_FULL`)**
   - Storage capacity exceeded
   - Typical Fix: Clear logs, free up space
   - Priority: **HIGH**

### Medium Priority
7. **Configuration Error (`config` / `CONFIG_ERROR`)**
   - Invalid configuration values detected
   - Typical Fix: Restore valid configuration and restart
   - Priority: **MEDIUM**

8. **Resource Exhaustion (`resource` / `RESOURCE_EXHAUSTION`)**
   - Ports or memory become unavailable
   - Typical Fix: Clear resources and restart
   - Priority: **MEDIUM**

9. **Memory Leak (`memory` / `MEMORY_LEAK`)**
   - Heap usage growing abnormally
   - Typical Fix: Restart service, investigate code
   - Priority: **MEDIUM**

### Low Priority
10. **High Latency (`latency` / `HIGH_LATENCY`)**
    - Response times exceeding threshold
    - Typical Fix: Optimize queries, scale resources
    - Priority: **LOW**

## 📊 Memory System

The system stores incident data in `incident_memory.json`:

```json
{
  "incidents": {
    "incident-id": {
      "id": "550e8400-...",
      "type": "SERVICE_DOWN",
      "status": "RESOLVED",
      "detected_at": "2025-01-15T10:30:00Z",
      "resolved_at": "2025-01-15T10:30:08Z",
      "resolution": {
        "fix_type": "restart",
        "steps": ["Stop service", "Start service", "Verify"],
        "success": true
      }
    }
  },
  "fixes": {
    "SERVICE_DOWN": {
      "fix_type": "restart",
      "steps": ["Stop service", "Start service", "Verify"]
    }
  }
}
```

## 🔧 Configuration

### Command Line Flags

- `-api-key string`: Cloudflare API key (defaults to `CLOUDFLARE_API_KEY` env var)
- `-account-id string`: Cloudflare Account ID (defaults to `CLOUDFLARE_ACCOUNT_ID` env var)
- `-use-ai bool`: Use Cloudflare AI for analysis (default: true)
- `-demo bool`: Run automated demo scenario (default: false)

### Environment Variables

- `CLOUDFLARE_API_KEY`: Your Cloudflare API key
- `CLOUDFLARE_ACCOUNT_ID`: Your Cloudflare Account ID

### Constants (in main.go)

- `servicePort`: Port for target service (default: "8080")
- `checkInterval`: Health check interval (default: 3 seconds)
- `memoryFile`: Path to incident memory file (default: "incident_memory.json")

## 🧪 Testing & Development

### Manual Testing

1. Start the system: `go run main.go`
2. Open another terminal
3. Trigger incidents manually using curl
4. Observe the logs to see detection → analysis → remediation → verification

### Automated Demo

Run the automated demo to see all incident types:

```bash
go run main.go -demo
```

This will:
1. Trigger a service crash
2. Trigger a config error
3. Trigger the same crash again (uses cached fix)
4. Trigger a dependency failure

### Fallback Mode

Test without Cloudflare AI credentials:

```bash
go run main.go -use-ai=false
```

The system uses rule-based logic as a fallback.

## 📝 Project Structure

```
incident-ai/
├── main.go                  # Entry point and orchestrator
├── go.mod                   # Go module definition
├── incident_memory.json     # Persistent storage (created at runtime)
├── models/
│   └── incident.go          # Core data structures
├── service/
│   └── target_service.go    # Simulated service with incident triggers
├── monitor/
│   └── detector.go          # Health monitoring and incident detection
├── ai/
│   └── analyzer.go          # Cloudflare AI integration and analysis
├── remediation/
│   └── executor.go          # Fix execution and service manipulation
└── memory/
    └── store.go             # Incident history and learned fixes
```

## 🎓 How It Works

### Detection Phase
1. Monitor polls service health every 3 seconds
2. When health check fails, creates an incident record
3. Analyzes symptoms to determine incident type

### Analysis Phase
1. Checks memory for previously learned fix
2. If found: Uses cached fix (fast path ⚡)
3. If not found: Calls Cloudflare AI with incident details
4. Cloudflare AI (Llama 3.3) returns diagnosis and fix steps

### Remediation Phase
1. Executor applies fix based on type:
   - **Restart**: Stops and starts the service
   - **Config**: Updates configuration and restarts
   - **Code**: Logs suggested code changes and restarts
2. Waits for service to stabilize

### Verification Phase
1. Runs 3 health checks with 1-second intervals
2. All must pass for incident to be marked resolved
3. Stores successful resolution in memory

### Learning Phase
1. Successful fixes are stored in memory
2. Next time same incident type occurs, cached fix is used
3. No AI call needed - instant resolution!

## 🔐 Security Notes

- API keys should be stored in environment variables, not committed to code
- The target service is for simulation only - not production-ready
- In production, you'd want authentication, rate limiting, and proper error handling

## 🐳 Docker Usage

### Building and Running with Docker

```bash
# Build the Docker image
docker build -t incident-ai .

# Run the container
docker run -p 8080:8080 \
  -e CLOUDFLARE_API_KEY=your-api-key \
  -e CLOUDFLARE_ACCOUNT_ID=your-account-id \
  incident-ai

# Or use docker-compose
docker-compose up -d

# View logs
docker-compose logs -f

# Stop the container
docker-compose down
```

### Environment Variables

- `CLOUDFLARE_API_KEY`: Your Cloudflare API key
- `CLOUDFLARE_ACCOUNT_ID`: Your Cloudflare Account ID
- `SERVICE_PORT`: Port for the service (default: 8080)

## 📈 Metrics & Analytics

### Basic Metrics

Access real-time metrics via HTTP:

```bash
curl http://localhost:8080/metrics
```

Returns JSON with:
- Total incidents handled
- Resolution success rate
- Incidents by type
- Learned fixes available

### Advanced Analytics & Trending

Access comprehensive analytics and trends:

```bash
curl http://localhost:8080/analytics
```

Returns detailed JSON report including:

**Summary Statistics:**
- Total incidents, resolved, failed
- Success rate percentage
- Resolution time statistics (avg, median, fastest, slowest)
- Cached fix usage rate
- AI calls made

**Trends:**
- Hourly incident trend with direction (increasing/decreasing/stable)
- Daily incident trend with change rate
- Per-type trends showing which incident types are growing or declining

**Hot Spots:**
- Most frequent incident type
- Most problematic type (highest failure rate)
- Time distribution (incidents by hour of day)

**Recent Activity:**
- Incidents in last 24 hours
- Incidents in last 7 days
- Incidents in last 30 days

**Predictions:**
- Predicted incidents for next hour (simple moving average)
- Overall trend status (improving/degrading/stable)

**Example analytics response:**
```json
{
  "total_incidents": 15,
  "resolved_incidents": 13,
  "failed_incidents": 2,
  "success_rate_percent": 86.67,
  "avg_resolution_time_seconds": 8.5,
  "median_resolution_time_seconds": 7.2,
  "fastest_resolution_seconds": 3.1,
  "slowest_resolution_seconds": 15.3,
  "cached_fix_usage_count": 7,
  "cached_fix_rate_percent": 53.85,
  "ai_calls_made": 6,
  "most_frequent_incident_type": "SERVICE_DOWN",
  "most_problematic_type": "DEPENDENCY_FAILURE",
  "incidents_last_24h": 15,
  "incidents_last_7d": 15,
  "incidents_last_30d": 15,
  "predicted_incidents_next_hour": 2.3,
  "trend_status": "stable",
  "hourly_trend": {
    "label": "Incidents per Hour",
    "points": [
      {"timestamp": "2025-01-15T10:00:00Z", "count": 3},
      {"timestamp": "2025-01-15T11:00:00Z", "count": 5},
      {"timestamp": "2025-01-15T12:00:00Z", "count": 7}
    ],
    "direction": "increasing",
    "change_rate_percent": 25.5
  },
  "type_trends": {
    "SERVICE_DOWN": {
      "label": "SERVICE_DOWN",
      "direction": "stable",
      "change_rate_percent": 0
    }
  }
}
```

## 🔧 Custom Remediation Scripts

Place custom scripts in the `scripts/` directory:

```bash
scripts/
├── SERVICE_DOWN.sh       # Handles service crashes
├── CONFIG_ERROR.sh       # Handles config issues
├── DATABASE_ERROR.bat    # Windows script for DB issues
└── README.md
```

Scripts receive the incident ID as the first argument and should exit with code 0 on success.

## ⏮️ Rollback Capabilities

The system automatically captures state before applying fixes:
- Configuration backup
- Service state snapshot
- Rollback steps generation

To manually trigger a rollback (requires code integration):
```go
err := executor.Rollback(incident)
```

## 🚀 Completed Features

- [x] Support for 10+ incident types
- [x] Incident prioritization (Critical/High/Medium/Low)
- [x] Metrics and analytics HTTP endpoint
- [x] Docker integration with docker-compose
- [x] Custom remediation scripts support
- [x] Rollback capabilities with state capture
- [x] Enhanced logging and monitoring
- [x] Advanced analytics and trending system
- [x] Hourly and daily trend analysis
- [x] Predictive incident forecasting
- [x] Hot spot identification

## 🚧 Future Enhancements

- [ ] Slack/email notifications
- [ ] Web dashboard for visualization
- [ ] Multi-service support
- [ ] Kubernetes integration
- [ ] Webhook support for external integrations
- [ ] Machine learning-based predictions

## 📄 License

This is a demo/educational project. Feel free to use and modify as needed!

## 🤝 Contributing

This is a demonstration project, but feel free to extend it for your own use cases!

## ❓ Troubleshooting

### "No Cloudflare API credentials provided"
- Set the `CLOUDFLARE_API_KEY` and `CLOUDFLARE_ACCOUNT_ID` environment variables, or
- Use the `-api-key` and `-account-id` flags, or
- Run with `-use-ai=false` for fallback mode

### "Port already in use"
- Stop any other processes using port 8080
- Or change `servicePort` in main.go

### "Health checks failing"
- Wait a few seconds for service to fully start
- Check if service is actually running on port 8080

### Memory file issues
- Delete `incident_memory.json` to start fresh
- Ensure write permissions in the directory

---


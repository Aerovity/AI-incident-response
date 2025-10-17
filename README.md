# 🤖 AI-Powered Incident Response System

An intelligent incident detection and auto-remediation system built in Go, inspired by incident.io's auto-remediation capabilities. The system uses OpenAI to analyze incidents and automatically apply fixes, while learning from past incidents to respond faster in the future.

## ✨ Features

- **🔍 Automatic Incident Detection**: Continuous health monitoring and incident detection
- **🤖 AI-Powered Analysis**: Uses OpenAI GPT-4 to diagnose root causes and suggest fixes
- **⚡ Smart Remediation**: Automatically applies fixes to resolve incidents
- **🧠 Learning System**: Remembers successful fixes and applies them instantly on recurrence
- **📊 Multiple Incident Types**: Handles service crashes, config errors, resource exhaustion, and dependency failures
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
3. **AI Analyzer** (`ai/`) - Integrates with OpenAI to analyze incidents and suggest fixes
4. **Remediation Executor** (`remediation/`) - Applies fixes to resolve incidents
5. **Memory Store** (`memory/`) - Stores incident history and learned fixes
6. **Models** (`models/`) - Core data structures

## 🚀 Quick Start

### Prerequisites

- Go 1.21 or higher
- OpenAI API key (optional - system works with fallback logic if not provided)

### Installation

1. Clone or download this project:
```bash
cd "C:\Users\House Computer\Desktop\AI incident Response"
```

2. Install dependencies:
```bash
go mod download
```

3. Set your OpenAI API key (optional):
```bash
# Windows
set OPENAI_API_KEY=sk-your-key-here

# Linux/Mac
export OPENAI_API_KEY=sk-your-key-here
```

### Running the System

**Basic mode (with OpenAI):**
```bash
go run main.go
```

**Without OpenAI (fallback mode):**
```bash
go run main.go -use-ai=false
```

**Automated demo:**
```bash
go run main.go -demo
```

**With explicit API key:**
```bash
go run main.go -api-key=sk-your-key-here
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
3. 🤖 If new: Ask OpenAI for diagnosis and fix
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

# Second time - uses cached fix (much faster!)
curl "http://localhost:8080/trigger-incident?type=crash"
```

The second incident will be resolved instantly using the learned fix!

### 4. Check Service Status

```bash
curl http://localhost:8080/status
```

### 5. View Summary

Press `Ctrl+C` to stop the system and see a summary of all incidents handled.

## 🎯 Example Output

```
[MONITOR] ⚠️  Health check FAILED - Incident detected!
══════════════════════════════════════════════════════════════════════
[DETECTOR] 🚨 Incident Detected: SERVICE_DOWN
[DETECTOR] ID: 550e8400-e29b-41d4-a716-446655440000
══════════════════════════════════════════════════════════════════════
[MEMORY] No cached fix found - using AI analysis
[AI] Calling OpenAI for incident analysis...
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

## �� Incident Types

### 1. Service Crash (`crash`)
- **Symptom**: Service stops responding to health checks
- **Typical Fix**: Restart the service
- **Use Case**: Process crashes, hangs, or becomes unresponsive

### 2. Configuration Error (`config`)
- **Symptom**: Invalid configuration values detected
- **Typical Fix**: Restore valid configuration and restart
- **Use Case**: Corrupted config files, invalid parameters

### 3. Resource Exhaustion (`resource`)
- **Symptom**: Resources (ports, memory) become unavailable
- **Typical Fix**: Clear resources and restart
- **Use Case**: Port conflicts, memory leaks, disk full

### 4. Dependency Failure (`dependency`)
- **Symptom**: External dependency (database) unreachable
- **Typical Fix**: Fix connection string and reconnect
- **Use Case**: Database down, API unavailable, network issues

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

- `-api-key string`: OpenAI API key (defaults to `OPENAI_API_KEY` env var)
- `-use-ai bool`: Use OpenAI for analysis (default: true)
- `-demo bool`: Run automated demo scenario (default: false)

### Environment Variables

- `OPENAI_API_KEY`: Your OpenAI API key

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

Test without OpenAI API key:

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
│   └── analyzer.go          # OpenAI integration and analysis
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
3. If not found: Calls OpenAI with incident details
4. OpenAI returns diagnosis and fix steps

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

## 🚀 Future Enhancements

- [ ] Support for more incident types
- [ ] Slack/email notifications
- [ ] Web dashboard for visualization
- [ ] Metrics and analytics
- [ ] Multi-service support
- [ ] Kubernetes integration
- [ ] Custom remediation scripts
- [ ] Incident prioritization
- [ ] Rollback capabilities

## 📄 License

This is a demo/educational project. Feel free to use and modify as needed!

## 🤝 Contributing

This is a demonstration project, but feel free to extend it for your own use cases!

## ❓ Troubleshooting

### "No OpenAI API key provided"
- Set the `OPENAI_API_KEY` environment variable, or
- Use the `-api-key` flag, or
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

## 📞 Support

For issues or questions, check the code comments or experiment with the system!

---

**Built with ❤️ and 🤖 by Claude**

# 🚀 Quick Setup Guide

This guide will help you get the AI-Powered Incident Response System up and running in minutes using Cloudflare Workers AI.

## Prerequisites

- **Go 1.21+**: [Download Go](https://golang.org/dl/)
- **Node.js & npm**: [Download Node.js](https://nodejs.org/)
- **Cloudflare Account**: [Sign up free](https://dash.cloudflare.com/sign-up)

## Step-by-Step Setup

### 1. Install Dependencies

```bash
# Install Go dependencies
go mod download

# Install Wrangler (Cloudflare CLI)
npm install -g wrangler
```

### 2. Authenticate with Cloudflare

```bash
# Login to Cloudflare (opens browser)
npx wrangler login
```

This command will:
1. Open your default browser
2. Ask you to sign in to Cloudflare
3. Request authorization for the app
4. Save credentials locally in `~/.wrangler/config/`

### 3. Run Setup Wizard (Optional)

```bash
# Automatically detect and configure your account
go run main.go -setup
```

The wizard will:
- ✓ Check if wrangler is installed
- ✓ Detect your Cloudflare Account ID
- ✓ Update `wrangler.toml` with your account

### 4. Start the System

```bash
# Run with Cloudflare AI
go run main.go

# Or run without AI (fallback mode)
go run main.go -use-ai=false
```

That's it! The system is now running on `http://localhost:8080`

## Testing the System

### Trigger an Incident

```bash
# In another terminal, trigger a service crash
curl "http://localhost:8080/trigger-incident?type=crash"
```

### Watch the Magic

The system will:
1. 🔍 Detect the unhealthy service
2. 🤖 Analyze with Cloudflare AI (Llama 3.3 70B)
3. 🔧 Apply the recommended fix
4. ✅ Verify the service is healthy again
5. 💾 Store the fix for future use

### Trigger the Same Incident Again

```bash
curl "http://localhost:8080/trigger-incident?type=crash"
```

This time it will use the **cached fix** - much faster! ⚡

### Check Analytics

```bash
# View comprehensive analytics
curl http://localhost:8080/analytics | jq

# View basic metrics
curl http://localhost:8080/metrics | jq
```

## Alternative Setup Methods

### Method 1: Wrangler (Recommended) ✅

```bash
npx wrangler login
go run main.go
```

**Pros:**
- ✅ Easiest setup
- ✅ Secure (credentials stored by Cloudflare)
- ✅ Same method as official Cloudflare projects

### Method 2: Environment Variables

```bash
# Windows
set CLOUDFLARE_API_KEY=your-api-key
set CLOUDFLARE_ACCOUNT_ID=your-account-id

# Linux/Mac
export CLOUDFLARE_API_KEY=your-api-key
export CLOUDFLARE_ACCOUNT_ID=your-account-id

go run main.go
```

**Pros:**
- ✅ Works without wrangler
- ✅ Good for CI/CD pipelines

**How to get credentials:**
1. Go to [Cloudflare Dashboard](https://dash.cloudflare.com/)
2. Click on "Workers & Pages" → "Overview"
3. Your Account ID is on the right side
4. For API Token: "My Profile" → "API Tokens" → "Create Token"
5. Use "Workers AI" template

### Method 3: No AI Mode

```bash
go run main.go -use-ai=false
```

**Pros:**
- ✅ No credentials needed
- ✅ Uses rule-based fallback logic
- ✅ Still functional for testing

## Configuration Files

### wrangler.toml

```toml
name = "incident-ai"
main = "main.go"
compatibility_date = "2025-01-15"

# Add your Account ID here (optional)
account_id = "your-account-id-from-wrangler-whoami"

[ai]
binding = "AI"
```

### .env (Optional)

```env
CLOUDFLARE_API_KEY=your-api-key-here
CLOUDFLARE_ACCOUNT_ID=your-account-id-here
```

## Verification Steps

### 1. Check Wrangler Installation

```bash
npx wrangler --version
# Should output: wrangler 3.x.x
```

### 2. Check Authentication

```bash
npx wrangler whoami
# Should show your email and account info
```

### 3. Test AI Connection

```bash
# Start the system
go run main.go

# Look for this message in the logs:
# ✓ Cloudflare credentials loaded successfully
```

### 4. Test Incident Response

```bash
# Trigger an incident
curl "http://localhost:8080/trigger-incident?type=crash"

# Watch the logs for AI analysis
# [AI] Calling Cloudflare AI for incident analysis...
# [AI] 📊 Diagnosis: ...
# [SYSTEM] ✅ INCIDENT RESOLVED!
```

## Common Issues & Solutions

### ❌ "Wrangler not found"

```bash
# Install wrangler globally
npm install -g wrangler

# Verify
npx wrangler --version
```

### ❌ "Cloudflare Authentication Not Found"

```bash
# Login to Cloudflare
npx wrangler login

# Verify authentication
npx wrangler whoami
```

### ❌ "Account ID not found"

**Option 1: Automatic**
```bash
go run main.go -setup
```

**Option 2: Manual**
```bash
# Get your account ID
npx wrangler whoami

# Add to wrangler.toml:
account_id = "abc123..."
```

### ❌ Port 8080 already in use

```bash
# Find and kill the process using port 8080
# Windows
netstat -ano | findstr :8080
taskkill /PID <process_id> /F

# Linux/Mac
lsof -i :8080
kill <process_id>
```

### ❌ "Module not found"

```bash
# Reinstall dependencies
go mod download
go mod tidy
```

## Running in Production

### Build the Binary

```bash
# Build
go build -o incident-ai

# Run
./incident-ai
```

### Using Docker

```bash
# Build image
docker build -t incident-ai .

# Run container
docker run -p 8080:8080 \
  -e CLOUDFLARE_API_KEY=your-key \
  -e CLOUDFLARE_ACCOUNT_ID=your-account \
  incident-ai

# Or with docker-compose
docker-compose up -d
```

## Next Steps

1. ✅ **Trigger different incident types** to see AI analysis
   ```bash
   curl "http://localhost:8080/trigger-incident?type=config"
   curl "http://localhost:8080/trigger-incident?type=dependency"
   ```

2. ✅ **Explore analytics** to see trends and patterns
   ```bash
   curl http://localhost:8080/analytics | jq
   ```

3. ✅ **Run the demo** to see full capabilities
   ```bash
   go run main.go -demo
   ```

4. ✅ **Read the docs** for advanced features
   - [README.md](README.md) - Full documentation
   - [ANALYTICS.md](ANALYTICS.md) - Analytics guide

## Getting Help

- **GitHub Issues**: Report bugs or request features
- **Cloudflare Docs**: [Workers AI Documentation](https://developers.cloudflare.com/workers-ai/)
- **Wrangler Docs**: [Wrangler Documentation](https://developers.cloudflare.com/workers/wrangler/)

## Free Tier Limits

Cloudflare Workers AI free tier includes:
- **10,000 AI requests per day**
- **Unlimited Worker invocations**
- Access to all AI models including Llama 3.3 70B

Perfect for development and testing! 🎉

---

**Happy incident hunting!** 🔥

For more information, see:
- [README.md](README.md) - Full documentation
- [ANALYTICS.md](ANALYTICS.md) - Analytics features

# ☁️ Cloudflare Authentication Guide

This project uses **Cloudflare Workers AI** for incident analysis, which provides free access to powerful AI models like Llama 3.3 70B.

## Why Cloudflare Workers AI?

✅ **Free Tier**: 10,000 AI requests per day
✅ **No Credit Card**: Free tier doesn't require payment info
✅ **Fast**: Edge computing for low latency
✅ **Powerful**: Access to Llama 3.3 70B (70 billion parameters)
✅ **Simple**: Just login with `npx wrangler login`

## Authentication Methods

### Method 1: Wrangler Login (Recommended) ⭐

This is the **easiest and most secure** method, used by official Cloudflare projects.

```bash
# Step 1: Install wrangler
npm install -g wrangler

# Step 2: Login to Cloudflare (opens browser)
npx wrangler login

# Step 3: Run setup wizard
go run main.go -setup

# Step 4: Start the system
go run main.go
```

**How it works:**
1. `npx wrangler login` opens your browser
2. You login to your Cloudflare account
3. Authorize the application
4. Credentials are stored securely in `~/.wrangler/config/`
5. The Go app reads credentials from wrangler's config

**Advantages:**
- ✅ Most secure (uses OAuth)
- ✅ No manual credential copying
- ✅ Same method as the [voice translator example](https://github.com/Aerovity/cloudflare-voice-translator)
- ✅ Credentials managed by Cloudflare's CLI

### Method 2: Environment Variables

If you prefer manual configuration or need it for CI/CD:

```bash
# Get your credentials from Cloudflare Dashboard
# 1. Go to: https://dash.cloudflare.com/
# 2. Click "Workers & Pages" → "Overview"
# 3. Account ID is on the right side
# 4. For API Token: "My Profile" → "API Tokens" → "Create Token"
# 5. Use "Workers AI" template

# Set environment variables
export CLOUDFLARE_API_KEY="your-api-token-here"
export CLOUDFLARE_ACCOUNT_ID="your-account-id-here"

# Run the system
go run main.go
```

**Advantages:**
- ✅ Good for CI/CD pipelines
- ✅ Works without Node.js/npm
- ✅ Explicit credential management

### Method 3: No AI Mode

For testing without Cloudflare:

```bash
go run main.go -use-ai=false
```

This uses built-in rule-based logic instead of AI.

## Comparison with Voice Translator Setup

This project uses the **same authentication approach** as the [Cloudflare Voice Translator](https://github.com/Aerovity/cloudflare-voice-translator):

| Voice Translator (Node.js) | Incident Response (Go) |
|----------------------------|------------------------|
| `npx wrangler login` | `npx wrangler login` ✅ |
| `npx wrangler deploy` | `go run main.go` |
| Workers + Pages | Standalone Go app + Workers AI API |
| KV storage for cache | File storage for incident memory |

**Key Difference:** This is a standalone Go application that **calls** Cloudflare Workers AI via REST API, rather than running **as** a Cloudflare Worker.

## Step-by-Step: First Time Setup

### 1. Create Cloudflare Account (if you don't have one)

Go to [cloudflare.com/sign-up](https://dash.cloudflare.com/sign-up) and create a free account.

### 2. Install Wrangler

```bash
# Install globally
npm install -g wrangler

# Verify installation
npx wrangler --version
# Output: wrangler 3.x.x
```

### 3. Login to Cloudflare

```bash
npx wrangler login
```

This will:
- Open `https://dash.cloudflare.com` in your browser
- Show an authorization prompt
- Ask you to click "Allow" to authorize wrangler
- Display "Successfully logged in" message

### 4. Verify Authentication

```bash
npx wrangler whoami
```

Output should show:
```
 ⛅️ wrangler 3.x.x
-------------------
Getting User settings...
👋 You are logged in with an OAuth Token, associated with the email 'your-email@example.com'!
┌──────────────────────┬──────────────────────────────────┐
│ Account Name         │ Account ID                       │
├──────────────────────┼──────────────────────────────────┤
│ Your Account         │ abc123def456...                  │
└──────────────────────┴──────────────────────────────────┘
```

### 5. Run Setup Wizard

```bash
go run main.go -setup
```

This automatically:
- ✓ Checks if wrangler is installed
- ✓ Gets your Account ID from `wrangler whoami`
- ✓ Updates `wrangler.toml` with your Account ID

### 6. Start the System

```bash
go run main.go
```

You should see:
```
✓ Cloudflare credentials loaded successfully
[SYSTEM] Starting target service...
```

## How Credentials Are Loaded

The application tries multiple sources in this order:

1. **Environment Variables** (highest priority)
   - `CLOUDFLARE_API_KEY`
   - `CLOUDFLARE_ACCOUNT_ID`

2. **Wrangler Config** (recommended)
   - Reads `wrangler.toml` for `account_id`
   - Reads `~/.wrangler/config/default.toml` for OAuth token

3. **Fallback Mode** (no credentials)
   - Uses rule-based logic instead of AI

## File Locations

### wrangler.toml (Project Directory)
```toml
name = "incident-ai"
main = "main.go"
compatibility_date = "2025-01-15"

# Your Account ID (from wrangler whoami)
account_id = "your-account-id"

[ai]
binding = "AI"
```

### ~/.wrangler/config/default.toml (User Home)
```toml
# OAuth token stored here after 'wrangler login'
oauth_token = "your-oauth-token"
```

**Note:** The Go app automatically reads both files, just like wrangler does!

## Troubleshooting

### ❌ "wrangler: command not found"

```bash
# Install Node.js first (includes npm)
# Download from: https://nodejs.org/

# Then install wrangler
npm install -g wrangler
```

### ❌ "Not logged in"

```bash
# Login again
npx wrangler login

# Verify
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

# Edit wrangler.toml and add:
account_id = "your-account-id-from-above"
```

### ❌ "API token invalid"

Your token might have expired. Re-login:

```bash
# Logout
npx wrangler logout

# Login again
npx wrangler login
```

### ❌ "Workers AI not available on free plan"

Workers AI **is** available on the free plan! Make sure:
- You're logged in: `npx wrangler whoami`
- Your account is verified (check email)
- Try logging out and in again

### ❌ Still having issues?

Run without AI to test the rest of the system:
```bash
go run main.go -use-ai=false
```

## Testing the Connection

### 1. Check Logs on Startup

```bash
go run main.go
```

Look for:
```
✓ Cloudflare credentials loaded successfully  ← Good!
```

Or:
```
⚠️  Cloudflare Authentication Not Found  ← Need to login
```

### 2. Trigger an Incident

```bash
curl "http://localhost:8080/trigger-incident?type=crash"
```

Watch the logs for:
```
[AI] Calling Cloudflare AI for incident analysis...  ← AI is working!
[AI] 📊 Diagnosis: Service process has crashed...
```

If you see this, Cloudflare AI is working! 🎉

### 3. Check Analytics

```bash
curl http://localhost:8080/analytics | jq '.ai_calls_made'
# Should be > 0 if AI was used
```

## Free Tier Limits

Cloudflare Workers AI free tier includes:

- ✅ **10,000 AI requests per day**
- ✅ **100,000 Worker requests per day**
- ✅ **Unlimited edge requests**
- ✅ **All AI models** (Llama 3.3 70B, Whisper, etc.)

**For this project:** Each incident uses ~1 AI request, so you can handle **10,000 incidents per day** for free! 🚀

## Security Notes

### Is it safe to use wrangler?

✅ **Yes!** Wrangler is Cloudflare's official CLI tool:
- Open source: [github.com/cloudflare/workers-sdk](https://github.com/cloudflare/workers-sdk)
- Uses OAuth (no passwords stored)
- Tokens stored locally in `~/.wrangler/`
- Used by thousands of developers

### Where are credentials stored?

- **Wrangler OAuth Token**: `~/.wrangler/config/default.toml`
- **Account ID**: `wrangler.toml` (in project directory)

Both are **local files**, not sent anywhere except Cloudflare's API.

### Can I use API keys instead?

Yes! Set environment variables:
```bash
export CLOUDFLARE_API_KEY="your-token"
export CLOUDFLARE_ACCOUNT_ID="your-id"
```

This is useful for:
- CI/CD pipelines
- Docker containers
- Environments without Node.js

## Next Steps

Once authenticated:

1. ✅ **Test incident response**
   ```bash
   curl "http://localhost:8080/trigger-incident?type=crash"
   ```

2. ✅ **Watch AI analysis** in the logs

3. ✅ **Check analytics**
   ```bash
   curl http://localhost:8080/analytics | jq
   ```

4. ✅ **Try different incident types** to see AI learning

## Resources

- [Cloudflare Workers AI Docs](https://developers.cloudflare.com/workers-ai/)
- [Wrangler Documentation](https://developers.cloudflare.com/workers/wrangler/)
- [Voice Translator Example](https://github.com/Aerovity/cloudflare-voice-translator)
- [Llama 3.3 70B Model](https://developers.cloudflare.com/workers-ai/models/llama-3.3-70b-instruct/)

---

**Questions?** Open an issue on GitHub or check the [SETUP.md](SETUP.md) guide!

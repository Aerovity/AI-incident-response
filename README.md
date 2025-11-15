# Incident Management Platform - Backend API

A multi-tenant incident management SaaS platform with AI-powered analysis and remediation, inspired by incident.io. Built with Go, Fiber, and Cloudflare AI.

## 🚀 Features

### Phase 1: Core Setup & Authentication ✅
- ✅ Multi-tenant architecture with organization isolation
- ✅ User authentication with JWT tokens
- ✅ Role-based access control (Admin, Responder, Member)
- ✅ JSON-based storage with file locking
- ✅ RESTful API with Fiber v2

### Phase 2: Incident Management (Coming Next)
- Incident CRUD operations
- Incident status workflow (Open → Investigating → Identified → Monitoring → Resolved)
- Incident assignment to users/teams
- Comments and updates on incidents
- Filtering by status, priority, assignment

### Phase 3: AI Integration (Planned)
- AI-powered incident analysis using Cloudflare Llama 3.3
- Automated remediation suggestions
- Learning from resolved incidents
- Confidence scoring

### Phase 4: Team Collaboration (Planned)
- Team management
- On-call scheduling and rotation
- Slack webhook notifications
- Email notifications (SMTP)

### Phase 5: Production Ready (Planned)
- Docker support
- Comprehensive documentation
- Health checks and metrics
- API rate limiting

## 📋 Prerequisites

- Go 1.21 or higher
- Git
- (Optional) Cloudflare account for AI features

## 🛠️ Installation

1. **Clone the repository**
   ```bash
   cd "c:\Users\House Computer\Desktop\AI incident Response"
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Configure environment variables**
   ```bash
   # Edit .env and set your configuration
   # IMPORTANT: Change JWT_SECRET to a secure random string (at least 32 characters)!
   ```

4. **Run the server**
   ```bash
   go run cmd/api/main.go
   ```

   The API will be available at `http://localhost:3000`

## 🔧 Configuration

Environment variables (`.env` file):

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `3000` |
| `JWT_SECRET` | JWT signing secret (min 32 chars) | **Required** |
| `JWT_EXPIRES_IN` | Token expiration time | `24h` |
| `CLOUDFLARE_API_KEY` | Cloudflare AI API key | Optional |
| `CLOUDFLARE_ACCOUNT_ID` | Cloudflare account ID | Optional |
| `ALLOWED_ORIGINS` | CORS allowed origins | `http://localhost:3001` |
| `DATA_DIR` | Data storage directory | `./data` |

## 📚 API Documentation

### Base URL
```
http://localhost:3000/api/v1
```

### Authentication Endpoints

#### 1. Register New User & Organization
Creates a new organization and the first admin user.

```bash
POST /api/v1/auth/register
```

**Request Body:**
```json
{
  "email": "admin@example.com",
  "name": "John Doe",
  "password": "SecurePass123",
  "organization_name": "Acme Inc",
  "organization_slug": "acme-inc"
}
```

**Response (201 Created):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "uuid-here",
    "email": "admin@example.com",
    "name": "John Doe",
    "role": "admin",
    "organization_id": "org-uuid",
    "created_at": "2025-01-15T10:00:00Z"
  },
  "organization": {
    "id": "org-uuid",
    "name": "Acme Inc",
    "slug": "acme-inc",
    "plan": "free",
    "auto_remediation": false,
    "created_at": "2025-01-15T10:00:00Z"
  }
}
```

**Validation Rules:**
- Email must be valid format
- Password must be at least 8 characters with 1 uppercase and 1 number
- Organization slug must be lowercase, alphanumeric with hyphens only

**Example with curl:**
```bash
curl -X POST http://localhost:3000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"admin@example.com\",\"name\":\"John Doe\",\"password\":\"SecurePass123\",\"organization_name\":\"Acme Inc\",\"organization_slug\":\"acme-inc\"}"
```

#### 2. Login
Authenticate and get JWT token.

```bash
POST /api/v1/auth/login
```

**Request Body:**
```json
{
  "email": "admin@example.com",
  "password": "SecurePass123",
  "organization_slug": "acme-inc"
}
```

**Response (200 OK):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "uuid-here",
    "email": "admin@example.com",
    "name": "John Doe",
    "role": "admin",
    "organization_id": "org-uuid",
    "created_at": "2025-01-15T10:00:00Z"
  },
  "organization": {
    "id": "org-uuid",
    "name": "Acme Inc",
    "slug": "acme-inc",
    "plan": "free",
    "auto_remediation": false
  }
}
```

**Example with curl:**
```bash
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"admin@example.com\",\"password\":\"SecurePass123\",\"organization_slug\":\"acme-inc\"}"
```

#### 3. Get Current User
Get authenticated user information.

```bash
GET /api/v1/auth/me
```

**Headers:**
```
Authorization: Bearer <your-jwt-token>
```

**Response (200 OK):**
```json
{
  "user": {
    "id": "uuid-here",
    "email": "admin@example.com",
    "name": "John Doe",
    "role": "admin",
    "organization_id": "org-uuid",
    "created_at": "2025-01-15T10:00:00Z"
  },
  "organization": {
    "id": "org-uuid",
    "name": "Acme Inc",
    "slug": "acme-inc",
    "plan": "free",
    "auto_remediation": false
  }
}
```

**Example with curl:**
```bash
# Save token from login/register
set TOKEN=your-jwt-token-here

curl -X GET http://localhost:3000/api/v1/auth/me -H "Authorization: Bearer %TOKEN%"
```

#### 4. Refresh Token
Refresh JWT token before expiration.

```bash
POST /api/v1/auth/refresh
```

**Headers:**
```
Authorization: Bearer <your-jwt-token>
```

**Response (200 OK):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**Example with curl:**
```bash
curl -X POST http://localhost:3000/api/v1/auth/refresh -H "Authorization: Bearer %TOKEN%"
```

### Health Check

```bash
GET /health
```

**Response (200 OK):**
```json
{
  "status": "ok",
  "time": "2025-01-15T10:00:00Z"
}
```

**Example with curl:**
```bash
curl http://localhost:3000/health
```

## 🧪 Testing Phase 1

Here's a complete test workflow:

### 1. Start the server
```bash
go run cmd/api/main.go
```

### 2. Test health check
```bash
curl http://localhost:3000/health
```

### 3. Register a new organization and user
```bash
curl -X POST http://localhost:3000/api/v1/auth/register -H "Content-Type: application/json" -d "{\"email\":\"admin@testorg.com\",\"name\":\"Test Admin\",\"password\":\"TestPass123\",\"organization_name\":\"Test Organization\",\"organization_slug\":\"test-org\"}"
```

Save the token from the response!

### 4. Login
```bash
curl -X POST http://localhost:3000/api/v1/auth/login -H "Content-Type: application/json" -d "{\"email\":\"admin@testorg.com\",\"password\":\"TestPass123\",\"organization_slug\":\"test-org\"}"
```

### 5. Get current user (with authentication)
```bash
# Replace TOKEN with your actual token
curl -X GET http://localhost:3000/api/v1/auth/me -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

### 6. Test without authentication (should fail)
```bash
curl -X GET http://localhost:3000/api/v1/auth/me
# Expected: 401 Unauthorized
```

### 7. Refresh token
```bash
curl -X POST http://localhost:3000/api/v1/auth/refresh -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

## 📁 Project Structure

```
incident-backend/
├── cmd/
│   └── api/
│       └── main.go              # Application entry point
├── internal/
│   ├── models/
│   │   ├── incident.go          # Incident models
│   │   ├── user.go              # User models
│   │   ├── organization.go      # Organization models
│   │   └── team.go              # Team models
│   ├── storage/
│   │   ├── json_store.go        # Generic JSON storage
│   │   ├── incident_store.go    # Incident persistence
│   │   ├── user_store.go        # User persistence
│   │   ├── org_store.go         # Organization persistence
│   │   └── team_store.go        # Team persistence
│   ├── handlers/
│   │   ├── auth.go              # Auth endpoints
│   │   ├── incidents.go         # Incident endpoints (Phase 2)
│   │   ├── teams.go             # Team endpoints (Phase 4)
│   │   └── users.go             # User endpoints
│   ├── services/
│   │   ├── auth_service.go      # Authentication logic
│   │   ├── incident_service.go  # Incident business logic (Phase 2)
│   │   ├── ai_service.go        # AI integration (Phase 3)
│   │   └── notification_service.go # Notifications (Phase 4)
│   ├── middleware/
│   │   ├── auth.go              # JWT authentication
│   │   ├── cors.go              # CORS configuration
│   │   └── logger.go            # Request logging
│   └── utils/
│       ├── password.go          # Password hashing
│       └── validators.go        # Input validation
├── data/                         # JSON data storage
│   ├── global/
│   │   └── organizations.json   # All organizations
│   └── organizations/
│       └── {org-id}/
│           ├── incidents.json
│           ├── users.json
│           ├── teams.json
│           └── learned_fixes.json
├── scripts/                      # Remediation scripts
├── .env                          # Environment configuration
├── .env.example                  # Example environment file
├── go.mod                        # Go dependencies
├── go.sum                        # Dependency checksums
└── README.md                     # This file
```

## 🔐 Security Features

- **Password Hashing**: bcrypt with cost 12
- **JWT Authentication**: HS256 signing with configurable expiry
- **CORS Protection**: Configurable allowed origins
- **Input Validation**: Email, password strength, required fields
- **Multi-Tenancy**: Organization-level data isolation
- **Role-Based Access**: Admin, Responder, Member roles

## 🐛 Error Handling

The API uses standard HTTP status codes:

- `200 OK`: Successful request
- `201 Created`: Resource created successfully
- `400 Bad Request`: Invalid request data
- `401 Unauthorized`: Missing or invalid authentication
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Server error

**Error Response Format:**
```json
{
  "error": "Error message here",
  "code": 400
}
```

## 🚧 Coming Next (Phase 2)

- Incident CRUD endpoints
- Incident status transitions
- Assignment to users/teams
- Comments and updates
- Filtering and pagination

## 📝 License

MIT License

## 👥 Contributing

Contributions welcome! Please open an issue or PR.

## 🔗 Links

- [Original Prototype](https://github.com/Aerovity/AI-incident-response)
- [Cloudflare AI](https://developers.cloudflare.com/workers-ai/)
- [Fiber Documentation](https://gofiber.io/)

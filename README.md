# Device Management Tool (DMT)

A Go-based REST API for managing devices with real-time notifications when employees have 3+ assigned devices.

## 🏗️ Project Structure

```
dmt/
├── main.go                     # Application entry point with graceful shutdown
├── internal/                   # Internal packages (not importable by external projects)
│   ├── app.go                 # HTTP server setup and routing
│   ├── db.go                  # Database connection management
│   ├── config/
│   │   └── env.go            # Environment configuration
│   ├── middleware/
│   │   └── keyauth.go        # API key authentication
│   └── migrations/           # Database schema migrations
├── pkg/device/               # Device domain package
│   ├── handler.go           # HTTP handlers for device operations
│   ├── db.go               # Database operations
│   ├── type.go             # Device data structures
│   ├── validation.go       # Input validation and sanitization
│   └── notify.go          # PostgreSQL listener for notifications
├── integration/            # Integration tests
├── docker-compose.yml     # Local development environment
└── Dockerfile            # Container build configuration
```

## 🚀 Quick Start

### Development

```bash
# Start services (PostgreSQL + Notification service)
docker-compose up -d
```

### API Usage

```bash
# Create device (requires Authorization: Bearer <base64-encoded-api-key>)
curl -X POST http://localhost:3000/api/v1/devices \
  -H "Authorization: Bearer <base64-key>" \
  -H "Content-Type: application/json" \
  -d '{"name":"Laptop","type":"laptop","ip":"192.168.1.100","mac":"aa:bb:cc:dd:ee:ff","employee":"jdo"}'

# List devices with filters
curl "http://localhost:3000/api/v1/devices?employee=jdo&type=laptop" \
  -H "Authorization: Bearer <base64-key>"
```

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run integration tests only
go test ./integration/...
```

## Decisions & Rationale

## Dependencies

I generally prefer to write more functionality myself and avoid pulling in unnecessary dependencies—especially given Go's excellent standard library, which often covers most needs out of the box. However, to keep development time low, I chose to use Fiber for its simplicity and performance, and pgx for its rich PostgreSQL support and better control over database interactions compared to the standard database/sql package.

### Familiar Structure

The project uses a structure that aligns with common Go practices. This reduces the onboarding time for new developers.
Database Choice

### PostgreSQL

is used instead of SQLite. While SQLite could have sufficed for the current scope, PostgreSQL aligns with Greenbone’s existing stack, enabling smoother integration and scalability.

### Notifications via Triggers

PostgreSQL triggers and listeners are used to separate route handling from device notification logic. This promotes a cleaner architecture and avoids the need for developers to manually manage notification logic on every update.

### Simplicity in Data Types

Native INET and MACADDR types were not used to avoid complexity, but they should be considered in a real-world system for data integrity and validation.

### Testing & DX

Significant effort went into integration testing because it's the most stable and valuable layer for ensuring system behavior. Good DX here leads to more thorough and confident testing. Live reloading of the DEV container would be also nice but skipped for now.

### Production Considerations

- For development speed, the API key is hardcoded. In production, keys should be stored in the db and hashed.

- Logging & Monitoring: Improve logging anbd add tracing, monitoring.

- Secrets Management: Use a secret manager instead of raw environment variables.

- Security: TLS should be enabled (ideally via a reverse proxy).

- Rate Limiting: Consider rate limiting depending on the deployment context.

- API Documentation: Should be added to improve developer usability and transparency.

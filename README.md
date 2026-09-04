# Survey Battle Backend - Go

A Go-based backend service for the Survey Battle game application. This API provides game logic, authentication, real-time features, and AI integration.

## Project Overview

This is a RESTful backend built with Go that handles:
- Game server operations
- User authentication and authorization
- Real-time communication
- Database operations (MongoDB)
- Caching (Redis)
- AI integration for game features
- JWT-based security

## Prerequisites

- **Go**: 1.26.5 or higher
- **MongoDB**: Running instance for data persistence
- **Redis**: Running instance for caching and real-time features
- **Environment Variables**: Configured `.env` file

## Project Structure

```
.
├── cmd/
│   └── server/
│       └── main.go           # Application entry point
├── internal/
│   ├── ai/                   # AI integration modules
│   ├── auth/                 # Authentication logic
│   ├── config/               # Configuration management
│   ├── game/                 # Game logic
│   ├── http/                 # HTTP handlers
│   ├── middleware/           # HTTP middleware
│   ├── mongo/                # MongoDB integration
│   ├── persistence/          # Data persistence layer
│   ├── realtime/             # Real-time communication
│   ├── redis/                # Redis integration
│   ├── repository/           # Repository pattern implementations
│   └── response/             # Response formatting utilities
├── tests/                    # Test files
├── go.mod                    # Go module definition
├── go.sum                    # Dependency checksums
├── .env                      # Environment variables (local)
├── .env.example              # Example environment variables
├── .air.toml                 # Air configuration for hot reload
└── README.md                 # This file
```

## Installation

### 1. Clone the Repository

```bash
git clone <repository-url>
cd survey-battle-backend-go
```

### 2. Install Dependencies

```bash
go mod download
go mod tidy
```

### 3. Setup Environment Variables

Copy the example environment file and configure it:

```bash
cp .env.example .env
```

Edit `.env` and fill in the required values:

```env
# Server Configuration
PORT=5000
NODE_ENV=development

# Database
MONGO_URI=mongodb://localhost:27017/survey-battle
REDIS_URL=redis://localhost:6379

# JWT Secrets (use strong random strings)
JWT_ACCESS_SECRET=your_access_secret_key
JWT_REFRESH_SECRET=your_refresh_secret_key
JWT_PLAYER_SECRET=your_player_secret_key

# JWT Expiration Times
JWT_ACCESS_EXPIRES_IN=15m
JWT_REFRESH_EXPIRES_IN=7d
JWT_PLAYER_EXPIRES_IN=30d

# Game Configuration
MIN_QUESTIONS_TO_PUBLISH=5

# LLM/AI Integration
LLM_PROVIDER=openai  # or another provider
LLM_MODEL=gpt-4
LLM_API_KEY=your_llm_api_key
```

## Running the Application

### Option 1: Direct Execution

```bash
go run ./cmd/server
```

The server will start on the port specified in `.env` (default: 5000).

### Option 2: Build and Run Binary

```bash
go build -o ./tmp/main ./cmd/server
./tmp/main
```

### Option 3: Development with Hot Reload (Recommended)

Install Air for automatic reloading on code changes:

```bash
go install github.com/cosmtrek/air@latest
```

Then run with hot reload:

```bash
air
```

The configuration for air is defined in `.air.toml` and will automatically rebuild the application when you make changes to Go files.

## API Endpoints

### Health Check

**GET** `/health`

Returns the health status of the API.

```bash
curl http://localhost:5000/health
```

Response:
```json
{
  "success": true,
  "message": "OK"
}
```

### Root Endpoint

**GET** `/`

Welcome message for the API.

```bash
curl http://localhost:5000/
```

Response:
```json
{
  "success": true,
  "message": "Family Feud Go backend"
}
```

## Development Workflow

### Running Tests

```bash
go test ./...
```

### Running Tests with Coverage

```bash
go test -cover ./...
```

### Code Formatting

```bash
go fmt ./...
```

### Code Linting

```bash
go vet ./...
```

### Building for Production

```bash
go build -o survey-battle-backend ./cmd/server
```

## Configuration Details

The application loads configuration from multiple sources (priority order):
1. Environment variables from `.env` file
2. System environment variables
3. Default values (if applicable)

Configuration is managed in `internal/config/config.go` and includes:
- Server port
- Node environment
- Database URIs (MongoDB, Redis)
- JWT secrets and expiration times
- AI provider settings
- Cookie and CORS settings

## Dependencies

Key dependencies managed in `go.mod`:
- `github.com/joho/godotenv` - Environment variable loading

## Database Setup

### MongoDB

Ensure MongoDB is running:

```bash
# Using Docker
docker run -d -p 27017:27017 --name mongodb mongo:latest
```

### Redis

Ensure Redis is running:

```bash
# Using Docker
docker run -d -p 6379:6379 --name redis redis:latest
```

## Security Notes

- **Never commit `.env` file** to version control
- Use strong random strings for JWT secrets
- Keep API keys and secrets secure
- Use HTTPS in production
- Validate all input data
- Implement rate limiting for API endpoints

## Troubleshooting

### Import Error: "could not import survey-battle-backend-go/internal/config"

Ensure the module name in `go.mod` matches the import paths in your code.

### Port Already in Use

Change the `PORT` environment variable to an available port:

```bash
PORT=8000 go run ./cmd/server
```

### Connection Errors

Verify MongoDB and Redis are running and accessible:

```bash
# Test MongoDB
mongo --eval "db.adminCommand('ping')"

# Test Redis
redis-cli ping
```

## Contributing

1. Create a feature branch
2. Make your changes
3. Format code: `go fmt ./...`
4. Run tests: `go test ./...`
5. Submit a pull request

## License

[Add your license here]

## Support

For issues, questions, or contributions, please open an issue on the repository.

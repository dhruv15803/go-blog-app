# Go Blog API Service

A RESTful API service built with Go and PostgreSQL for a blogging platform with user authentication, blog management, comments, and topic organization.

## Table of Contents

- [Features](#features)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Configuration](#configuration)
- [Database Setup](#database-setup)
- [Running the Service](#running-the-service)
- [API Documentation](#api-documentation)
- [Project Structure](#project-structure)
- [Testing](#testing)
- [Docker Support](#docker-support)
- [Environment Variables](#environment-variables)
- [Contributing](#contributing)
- [License](#license)

## Features

- User registration and authentication with email activation
- JWT-based authentication system
- Blog post creation, management, and status updates
- **Intelligent blog feed algorithm** with personalized content ranking
- Blog liking and bookmarking functionality
- Hierarchical commenting system with nested replies
- Topic-based blog organization and following
- Admin-protected routes for topic management
- Email notifications via SMTP
- Health check endpoints
- Middleware for authentication and authorization
- Structured logging with Chi router
- Database connection pooling
- Graceful error handling

## Prerequisites

- Go 1.21 or higher
- PostgreSQL 13 or higher
- Git

## Installation

1. Clone the repository:
```bash
git clone https://github.com/dhruv15803/go-blog-app.git
cd go-blog-app
```

2. Install Go dependencies:
```bash
go mod download
```

3. Install development tools (optional):
```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/swaggo/swag/cmd/swag@latest
```

## Configuration

The service uses environment variables for configuration. Create a `.env` file in the root directory with your configuration values.

## Database Setup

1. Create a PostgreSQL database:
```sql
CREATE DATABASE blog_db;
CREATE USER blog_user WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE blog_db TO blog_user;
```

2. Run database migrations (if you have migration files):
```bash
# Using golang-migrate tool
migrate -path ./migrations -database "postgres://blog_user:your_password@localhost:5432/blog_db?sslmode=disable" up
```

3. Your connection string should look like:
```
POSTGRES_DB_CONN=postgres://blog_user:your_password@localhost:5432/blog_db?sslmode=disable
```

## Running the Service

### Development Mode

```bash
go run cmd/api/main.go cmd/api/api.go cmd/api/db.go
```

### Production Mode

1. Build the binary:
```bash
go build -o bin/blog-api cmd/api/main.go cmd/api/api.go cmd/api/db.go
```

2. Run the binary:
```bash
./bin/blog-api
```

The service will start on the port specified in your configuration (default: 8080).

## API Documentation

### Base URL
```
http://localhost:{PORT}/api
```

### Health Check
```
GET /health
```

### Authentication Endpoints
```
POST /auth/register           # Register a new user
PUT  /auth/activate/{token}   # Activate user account via email token
POST /auth/login              # User login
GET  /auth/user               # Get authenticated user info (requires auth)
```

### Blog Endpoints
```
GET    /blog/blogs/feed                    # Get personalized blog feed (optional auth)
GET    /blog/{topicId}/blogs               # Get blogs by topic (public)
POST   /blog/                              # Create a new blog post (requires auth)
DELETE /blog/{blogId}                      # Delete a blog post (requires auth)
PATCH  /blog/{blogId}/status               # Update blog status (requires auth)
POST   /blog/{blogId}/like                 # Like/unlike a blog post (requires auth)
POST   /blog/{blogId}/bookmark             # Bookmark/unbookmark a blog (requires auth)
```

#### Blog Feed Algorithm
The `/blog/blogs/feed` endpoint implements an intelligent content ranking system:

**For Authenticated Users:**
- Fetches blogs from topics the user follows
- Ranks content using activity score algorithm
- Returns paginated, personalized feed

**For Unauthenticated Users:**
- Fetches blogs from the top 5 most-followed topics
- Applies same activity ranking algorithm
- Returns paginated general feed

**Activity Score Formula:**
```
activity_score = (0.3 × likes + 0.5 × comments + 0.2 × bookmarks) / (time_since_creation_minutes)²
```

This algorithm prioritizes:
- Recent content (time decay factor)
- High engagement (comments weighted highest)
- Community interaction (likes and bookmarks)

### Blog Comment Endpoints
```
GET    /blog-comment/{blogId}/blog-comments          # Get comments for a blog
GET    /blog-comment/{blogCommentId}/comments        # Get replies to a comment
POST   /blog-comment/                                # Create a new comment (requires auth)
DELETE /blog-comment/{blogCommentId}                 # Delete a comment (requires auth)
PUT    /blog-comment/{blogCommentId}                 # Update a comment (requires auth)
POST   /blog-comment/{blogCommentId}/like            # Like/unlike a comment (requires auth)
```

### Topic Endpoints
```
GET    /topic/topics                       # Get all topics (public)
POST   /topic/                             # Create a topic (admin only)
PUT    /topic/{topicId}                    # Update a topic (admin only)
DELETE /topic/{topicId}                    # Delete a topic (admin only)
POST   /topic/{topicId}/follow             # Follow/unfollow a topic (requires auth)
```

### Example Request/Response

**POST /api/auth/register**
```json
{
  "email": "user@example.com",
  "password": "securepassword123",
  "username": "johndoe"
}
```

**Response (201 Created):**
```json
{
  "message": "User registered successfully. Please check your email for activation."
}
```

**POST /api/blog/**
```json
{
  "title": "My First Blog Post",
  "content": "This is the content of my blog post...",
  "topic_id": 1,
  "status": "published"
}
```

**Response (201 Created):**
```json
{
  "id": 123,
  "title": "My First Blog Post",
  "content": "This is the content of my blog post...",
  "topic_id": 1,
  "status": "published",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

For detailed API documentation, visit `/swagger/index.html` when running the service (if Swagger is integrated).

## Project Structure

```
.
├── bin/                       # Build output directory
├── cmd/
│   ├── api/
│   │   ├── api.go            # Server setup and route definitions
│   │   ├── db.go             # Database connection configuration  
│   │   └── main.go           # Application entry point with config loading
│   └── createUser/
│       └── main.go           # User creation utility
├── internal/
│   ├── handlers/             # HTTP request handlers
│   ├── mailer/               # Email service functionality
│   └── storage/              # Data access layer / repositories
├── migrations/               # Database migration files
├── scripts/
│   └── user.go              # User-related scripts
├── templates/                # Email/HTML templates
├── utils/                    # Utility functions
├── .env                      # Environment variables (not committed)
├── .gitignore               # Git ignore rules
├── go.mod                   # Go module dependencies
├── go.sum                   # Go module checksums
└── README.md                # This documentation
```

## Testing

### Run All Tests
```bash
go test ./...
```

### Run Tests with Coverage
```bash
go test -v -cover ./...
```

### Run Integration Tests
```bash
go test -tags=integration ./tests/integration/...
```

### Run Specific Test Package
```bash
go test ./internal/service/...
```

## Docker Support

*Note: Docker configuration files are not currently included in this project. If you'd like to add Docker support, you can create the following files:*

### Create a Dockerfile:
```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main cmd/api/main.go cmd/api/api.go cmd/api/db.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/.env .

EXPOSE 8080
CMD ["./main"]
```

### Create a docker-compose.yml:
```yaml
version: '3.8'
services:
  api:
    build: .
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
      - POSTGRES_DB_CONN=postgres://blog_user:password@db:5432/blog_db?sslmode=disable
      - JWT_SECRET=your-jwt-secret
      - CLIENT_URL=http://localhost:3000
    depends_on:
      - db
    env_file:
      - .env

  db:
    image: postgres:13
    environment:
      - POSTGRES_DB=blog_db
      - POSTGRES_USER=blog_user
      - POSTGRES_PASSWORD=password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

Then you can use:
```bash
docker-compose up -d
```

## Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `PORT` | Server port | | Yes |
| `POSTGRES_DB_CONN` | PostgreSQL connection string | | Yes |
| `MAILER_HOST` | SMTP server host | | Yes |
| `MAILER_PORT` | SMTP server port | | Yes |
| `MAILER_USERNAME` | SMTP username | | Yes |
| `MAILER_PASSWORD` | SMTP password | | Yes |
| `CLIENT_URL` | Frontend application URL | | Yes |
| `JWT_SECRET` | JWT signing secret | | Yes |
| `GO_ENV` | Environment (development, staging, production) | `development` | No |

### Example .env file:
```env
PORT=8080
POSTGRES_DB_CONN=postgres://username:password@localhost:5432/blog_db?sslmode=disable
MAILER_HOST=smtp.gmail.com
MAILER_PORT=587
MAILER_USERNAME=your-email@gmail.com
MAILER_PASSWORD=your-app-password
CLIENT_URL=http://localhost:3000
JWT_SECRET=your-super-secret-jwt-key-min-32-chars
GO_ENV=development
```

## Build Commands

Since this project doesn't include a Makefile, here are the common commands you can use:

```bash
# Run the application in development
go run cmd/api/main.go cmd/api/api.go cmd/api/db.go

# Build the application
go build -o bin/blog-api cmd/api/main.go cmd/api/api.go cmd/api/db.go

# Run tests
go test -v ./...

# Run tests with coverage
go test -v -cover ./...

# Run linter (if golangci-lint is installed)
golangci-lint run

# Clean build artifacts
rm -rf bin/

# Tidy up dependencies
go mod tidy

# Download dependencies
go mod download
```

### Optional: Create a Makefile

If you'd like to add a Makefile for easier command management, create this file in your root directory:

```makefile
.PHONY: build run test lint clean

# Build the application
build:
	go build -o bin/blog-api cmd/api/main.go cmd/api/api.go cmd/api/db.go

# Run the application
run:
	go run cmd/api/main.go cmd/api/api.go cmd/api/db.go

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -cover ./...

# Run linter
lint:
	golangci-lint run

# Clean build artifacts
clean:
	rm -rf bin/

# Tidy dependencies
tidy:
	go mod tidy
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow Go best practices and conventions
- Write tests for new functionality
- Update documentation as needed
- Run `make lint` before submitting PRs
- Ensure all tests pass

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

If you have any questions or run into issues, please:

1. Check the existing issues on GitHub
2. Create a new issue with detailed information
3. Contact the development team at [your-email@example.com]

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for a detailed history of changes.
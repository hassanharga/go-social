# Go-Social

A robust social networking API built with Go, featuring user authentication, post management, commenting system, follower relationships, and user feeds.

## Features

### User Management
- **User Registration & Authentication**: Secure user registration with email verification
- **JWT Authentication**: Token-based authentication for secure API access
- **Role-based Access Control**: Different permission levels (user, moderator, admin)
- **User Activation**: Email-based account activation system
- **User Profiles**: View and manage user profiles
- **Follow System**: Users can follow/unfollow other users

### Content Management
- **Posts**: Create, read, update, and delete posts
- **Comments**: Add comments to posts
- **Tag System**: Tag posts with relevant keywords
- **Search**: Search posts by title, content, or tags
- **Feed**: Personalized feed based on followed users

### Technical Features
- **Rate Limiting**: Prevent API abuse with configurable rate limiting
- **Caching**: Redis-based caching for improved performance
- **Database Migrations**: Structured database schema management
- **Email Integration**: Support for SendGrid and Mailtrap email services
- **API Documentation**: Swagger/OpenAPI documentation
- **Error Handling**: Comprehensive error handling and logging
- **Testing**: Unit and integration tests

## Tech Stack

- **Backend**: Go
- **Database**: PostgreSQL
- **Caching**: Redis
- **API Documentation**: Swagger
- **Email Services**: SendGrid/Mailtrap
- **Authentication**: JWT
- **Containerization**: Docker

## Project Structure

```
├── cmd/                   # Application entry points
│   ├── api/               # Main API server
│   └── migrate/           # Database migrations
├── internal/              # Internal packages
│   ├── auth/              # Authentication logic
│   ├── db/                # Database connection and utilities
│   ├── env/               # Environment configuration
│   ├── mailer/            # Email service integration
│   ├── ratelimiter/       # API rate limiting
│   └── store/             # Data access layer
├── docs/                  # API documentation
├── scripts/               # Utility scripts
└── utils/                 # Helper utilities
```

## Getting Started

### Prerequisites

- Go 1.24+
- PostgreSQL
- Redis
- Docker & Docker Compose (optional)

### Environment Variables

Create a `.env` file in the root directory with the following variables:

```env
# Server
PORT=3000
ENV=development
FRONTEND_URL=http://localhost:3001

# Database
DB_ADDR=postgres://admin:adminpassword@localhost/social?sslmode=disable
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=25
DB_MAX_IDLE_TIME=15m

# Redis
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# Authentication
JWT_SECRET=your_secret_key
JWT_AUD=goSocial
JWT_ISS=goSocial

# Basic Auth (for admin endpoints)
BASIC_AUTH_USER=admin
BASIC_AUTH_PASSWORD=adminpassword

# Email
FROM_EMAIL=noreply@localhost
SENDGRID_API_KEY=your_sendgrid_api_key
MAILTRAP_API_KEY=your_mailtrap_api_key
```

### Running with Docker

```bash
# Build and start containers
docker-compose up -d
```

### Running Locally

```bash
# Run migrations
make migrate-up

# Seed database (optional)
make seed

# Start the API server with go-air
air
```

### API Documentation

Once the server is running, access the Swagger documentation at:
```
http://localhost:3000/swagger/index.html
```

## Testing

```bash
# Run all tests
make test

# Run API tests
# go test ./cmd/api/...
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.
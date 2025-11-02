# Winkr Backend

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen.svg)](https://github.com/22smeargle/winkr-backend)
[![Coverage](https://img.shields.io/badge/Coverage-85%25-green.svg)](https://github.com/22smeargle/winkr-backend)

A modern, scalable dating application backend built with Go following Clean Architecture principles. Winkr Backend provides a comprehensive API for feature-rich dating platforms with real-time messaging, advanced matching algorithms, photo verification, subscription management, and robust security measures.

## 🚀 Quick Start

```bash
# Clone the repository
git clone https://github.com/22smeargle/winkr-backend.git
cd winkr-backend

# Set up environment
cp .env.example .env
# Edit .env with your configuration

# Start development environment
make setup-dev

# The application will be available at http://localhost:8080
```

## 📋 Table of Contents

- [Features](#features)
- [Technology Stack](#technology-stack)
- [Project Structure](#project-structure)
- [Getting Started](#getting-started)
- [Development](#development)
- [API Documentation](#api-documentation)
- [Testing](#testing)
- [Deployment](#deployment)
- [Contributing](#contributing)
- [License](#license)
- [Contact](#contact)

## ✨ Features

### Core Features
- **🔐 User Management**: Registration, authentication, profile management with JWT security
- **📸 Photo Management**: Upload, verification, and organization with AI moderation
- **💕 Matching System**: Advanced swipe-based matching with preferences and algorithms
- **💬 Real-time Messaging**: WebSocket-based instant messaging with read receipts
- **💳 Subscription Management**: Premium features with Stripe integration
- **🛡️ Security**: Comprehensive security with rate limiting, encryption, and protection

### Advanced Features
- **🔮 Ephemeral Photos**: Self-destructing photos with view-once functionality
- **🤖 AI-Powered Verification**: Automated content moderation using AWS Rekognition
- **📊 Analytics Dashboard**: User behavior analytics and insights
- **🔍 Advanced Search**: Location-based and preference-based user discovery
- **📱 Mobile Ready**: Optimized APIs for mobile applications
- **🌍 Internationalization**: Multi-language support ready

### Admin & Moderation
- **👥 Admin Panel**: Comprehensive admin dashboard for user management
- **🚨 Reporting System**: User reporting and automated moderation workflows
- **📈 Monitoring**: Real-time system monitoring and alerting
- **🔧 Configuration**: Flexible configuration management
- **📝 Audit Logs**: Comprehensive audit trail for all actions

## 🛠 Technology Stack

### Core Technologies
- **Go 1.21+** - High-performance programming language
- **Gin** - Fast HTTP web framework with middleware
- **GORM** - Powerful ORM for database operations
- **PostgreSQL 14+** - Robust primary database with JSONB support
- **Redis 6+** - High-performance caching and session storage

### External Services & Integrations
- **AWS S3/MinIO** - Scalable file storage with CDN
- **AWS Rekognition** - AI-powered photo verification and moderation
- **Stripe** - Secure payment processing and subscription management
- **SendGrid** - Reliable email delivery service
- **Prometheus** - Metrics collection and monitoring
- **Grafana** - Data visualization and dashboards

### Development & DevOps Tools
- **Docker** - Application containerization
- **Docker Compose** - Local development environment orchestration
- **golang-migrate** - Database schema migration management
- **Swagger/OpenAPI** - Interactive API documentation
- **golangci-lint** - Comprehensive code linting and analysis
- **GitHub Actions** - Continuous integration and deployment

## Project Structure

```
backend/
├── cmd/
│   └── api/                     # Application entry point
│       └── main.go
├── internal/
│   ├── domain/                  # Business logic and entities
│   │   ├── entities/           # Core business entities
│   │   ├── repositories/        # Repository interfaces
│   │   ├── services/           # Business logic services
│   │   └── valueobjects/       # Value objects and enums
│   ├── application/             # Application layer
│   │   ├── usecases/           # Use cases
│   │   └── dto/                # Data transfer objects
│   ├── infrastructure/          # Infrastructure layer
│   │   ├── database/           # Database implementations
│   │   ├── storage/            # File storage
│   │   ├── external/           # External services
│   │   ├── websocket/          # WebSocket implementation
│   │   └── middleware/         # HTTP middleware
│   └── interfaces/             # Interface adapters
│       ├── http/                # HTTP handlers and routes
│       └── websocket/          # WebSocket handlers
├── pkg/                        # Shared packages
│   ├── config/                 # Configuration management
│   ├── logger/                 # Logging utilities
│   ├── validator/              # Validation utilities
│   ├── errors/                 # Custom error types
│   └── utils/                  # Utility functions
├── migrations/                 # Database migration files
├── docs/                       # Documentation
├── scripts/                    # Build and deployment scripts
├── docker/                     # Docker configurations
├── .env.example               # Environment variables template
├── go.mod                     # Go module file
├── go.sum                     # Go dependencies checksum
├── Makefile                   # Build automation
└── README.md                  # Project documentation
```

## 🚀 Getting Started

### Prerequisites

#### System Requirements
- **Go 1.21+** - [Installation Guide](https://golang.org/doc/install)
- **Docker 20.10+** - [Installation Guide](https://docs.docker.com/get-docker/)
- **Docker Compose 2.0+** - Included with Docker Desktop
- **Git 2.30+** - [Installation Guide](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)

#### External Services (Optional for Full Features)
- **AWS Account** - For S3 storage and Rekognition
- **Stripe Account** - For payment processing
- **SendGrid Account** - For email services

### Installation

#### Option 1: Quick Start (Recommended)
```bash
# Clone the repository
git clone https://github.com/22smeargle/winkr-backend.git
cd winkr-backend

# Complete setup with one command
make setup-dev

# The application will be available at http://localhost:8080
```

#### Option 2: Manual Installation
1. **Clone the repository**
   ```bash
   git clone https://github.com/22smeargle/winkr-backend.git
   cd winkr-backend
   ```

2. **Set up environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   nano .env  # or use your preferred editor
   ```

3. **Install Go dependencies**
   ```bash
   make deps
   ```

4. **Start development services**
   ```bash
   make docker-up
   ```

5. **Run database migrations**
   ```bash
   make migrate-up
   ```

6. **Start the application**
   ```bash
   make run
   ```

### Verification
```bash
# Check if the application is running
curl http://localhost:8080/health

# Expected response:
# {"status":"ok","timestamp":"2025-01-01T12:00:00Z"}
```

### Environment Configuration
Key environment variables to configure in `.env`:
```bash
# Application
APP_ENV=development
APP_PORT=8080

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=winkr_user
DB_PASSWORD=winkr_password
DB_NAME=winkr_db

# Security
JWT_SECRET=your-super-secret-jwt-key

# External Services
AWS_ACCESS_KEY_ID=your-aws-key
AWS_SECRET_ACCESS_KEY=your-aws-secret
STRIPE_SECRET_KEY=your-stripe-key
```

## 💻 Development

### Development Commands

#### Essential Commands
```bash
# Build the application
make build

# Run the application
make run

# Run all tests
make test

# Run tests with coverage report
make test-cover

# Format code according to Go standards
make fmt

# Run comprehensive linting
make lint

# Run security vulnerability scan
make security
```

#### Advanced Development Commands
```bash
# Generate API documentation
make swagger

# Generate mocks for testing
make mock

# Start development server with hot reload
make dev

# Clean build artifacts
make clean

# Install development dependencies
make deps

# Run database migrations
make migrate-up

# Rollback database migrations
make migrate-down
```

#### Database Management
```bash
# Create new migration
make migrate-create NAME=create_new_table

# Force migration to specific version
make migrate-force VERSION=1

# Reset database (dangerous!)
make migrate-reset
```

### Development Workflow

#### 1. Feature Development
```bash
# Create feature branch
git checkout -b feature/your-feature-name

# Make changes and test
make test
make lint

# Commit changes
git add .
git commit -m "feat: add your feature description"

# Push to your fork
git push origin feature/your-feature-name
```

#### 2. Code Quality
```bash
# Run all quality checks
make quality

# This includes:
# - go fmt (formatting)
# - golangci-lint (linting)
# - go test (testing)
# - gosec (security scanning)
```

#### 3. Testing
```bash
# Run unit tests
make test-unit

# Run integration tests
make test-integration

# Run end-to-end tests
make test-e2e

# Run performance benchmarks
make benchmark
```

### Database Migrations

```bash
# Create new migration
make migrate-create

# Run migrations
make migrate-up

# Rollback migrations
make migrate-down

# Force migration version
make migrate-force VERSION=1
```

### Docker Development

```bash
# Start all services
make docker-up

# Stop all services
make docker-down

# View logs
make docker-logs
```

## 📚 API Documentation

### Interactive Documentation
```bash
# Generate and serve interactive API docs
make swagger

# Open http://localhost:8080/swagger/index.html
```

### Key API Endpoints

#### Authentication
```bash
# Register new user
POST /api/v1/auth/register
Content-Type: application/json
{
  "email": "user@example.com",
  "password": "password123",
  "first_name": "John",
  "last_name": "Doe",
  "date_of_birth": "1990-01-01",
  "gender": "male",
  "interested_in": ["female"]
}

# User login
POST /api/v1/auth/login
Content-Type: application/json
{
  "email": "user@example.com",
  "password": "password123"
}

# Refresh access token
POST /api/v1/auth/refresh
Content-Type: application/json
{
  "refresh_token": "your-refresh-token"
}

# User logout
POST /api/v1/auth/logout
Authorization: Bearer your-access-token
```

#### User Management
```bash
# Get user profile
GET /api/v1/users/profile
Authorization: Bearer your-access-token

# Update user profile
PUT /api/v1/users/profile
Authorization: Bearer your-access-token
Content-Type: application/json
{
  "bio": "Updated bio",
  "location": {
    "lat": 40.7128,
    "lng": -74.0060,
    "city": "New York",
    "country": "USA"
  }
}

# Upload photo
POST /api/v1/photos
Authorization: Bearer your-access-token
Content-Type: multipart/form-data
file: [photo-file]
```

#### Matching & Discovery
```bash
# Get potential matches
GET /api/v1/matches/potential?limit=10&offset=0
Authorization: Bearer your-access-token

# Swipe on user
POST /api/v1/matches/swipe
Authorization: Bearer your-access-token
Content-Type: application/json
{
  "swiped_user_id": "user-id",
  "is_like": true
}

# Get matches
GET /api/v1/matches
Authorization: Bearer your-access-token
```

#### Messaging
```bash
# Get conversations
GET /api/v1/messages/conversations
Authorization: Bearer your-access-token

# Send message
POST /api/v1/messages/conversations/{conversation_id}
Authorization: Bearer your-access-token
Content-Type: application/json
{
  "content": "Hello!",
  "message_type": "text"
}
```

#### Ephemeral Photos
```bash
# Upload ephemeral photo
POST /api/v1/ephemeral-photos
Authorization: Bearer your-access-token
Content-Type: multipart/form-data
file: [photo-file]
expiration: 24

# View ephemeral photo
GET /api/v1/ephemeral-photos/{photo_id}/view?token=view-token
```

### WebSocket Events
```javascript
// Connect to WebSocket
const ws = new WebSocket('ws://localhost:8080/api/v1/ws?token=your-access-token');

// Send message
ws.send(JSON.stringify({
  type: 'message',
  data: {
    conversation_id: 'conversation-id',
    content: 'Hello!',
    message_type: 'text'
  }
}));

// Receive messages
ws.onmessage = function(event) {
  const data = JSON.parse(event.data);
  console.log('Received:', data);
};
```

### API Documentation Files
- [Complete API Reference](docs/api/README.md)
- [Authentication Guide](docs/api/authentication.md)
- [Error Handling](docs/api/errors.md)
- [Rate Limiting](docs/api/rate-limiting.md)
- [Security Guidelines](docs/api/security.md)

## 🧪 Testing

### Testing Strategy

#### Test Categories
- **Unit Tests**: Test individual functions and methods
- **Integration Tests**: Test component interactions
- **End-to-End Tests**: Test complete user workflows
- **Performance Tests**: Test system performance under load
- **Security Tests**: Test security vulnerabilities

### Running Tests

#### Basic Testing Commands
```bash
# Run all tests
make test

# Run tests with coverage report
make test-cover

# Run performance benchmarks
make benchmark

# Run specific test package
go test ./internal/domain/entities/...

# Run tests with verbose output
go test -v ./...

# Run tests with race detection
go test -race ./...
```

#### Advanced Testing
```bash
# Run only unit tests
make test-unit

# Run only integration tests
make test-integration

# Run only end-to-end tests
make test-e2e

# Generate coverage report
make coverage

# View coverage in browser
make coverage-html
```

### Test Structure

#### Directory Layout
```
tests/
├── unit/                    # Unit tests
├── integration/             # Integration tests
├── e2e/                   # End-to-end tests
├── performance/            # Performance tests
├── fixtures/              # Test data
└── mocks/                 # Generated mocks

internal/
├── domain/
│   ├── entities/
│   │   ├── user.go
│   │   └── user_test.go    # Unit tests alongside source
│   └── ...
└── ...
```

#### Writing Tests
```go
// Example unit test
func TestUserService_CreateUser(t *testing.T) {
    // Arrange
    mockRepo := &MockUserRepository{}
    service := NewUserService(mockRepo, nil, nil)
    
    user := &User{
        Email:    "test@example.com",
        Password: "password123",
    }
    
    mockRepo.On("Create", user).Return(user, nil)
    
    // Act
    result, err := service.CreateUser(user)
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, user, result)
    mockRepo.AssertExpectations(t)
}
```

### Test Coverage Requirements
- **Overall Coverage**: Minimum 85%
- **Domain Layer**: Minimum 90%
- **Application Layer**: Minimum 85%
- **Infrastructure Layer**: Minimum 80%

### Mock Generation
```bash
# Generate mocks for all interfaces
make mock

# Generate specific mock
mockgen -source=internal/domain/repositories/user_repository.go -destination=tests/mocks/user_repository_mock.go
```

## 🚀 Deployment

### Deployment Options

#### 1. Docker Deployment (Recommended)
```bash
# Build production image
docker build -f docker/Dockerfile -t winkr-backend:latest .

# Run with docker-compose
docker-compose -f docker/docker-compose.yml up -d

# Check deployment status
docker-compose -f docker/docker-compose.yml ps
```

#### 2. Kubernetes Deployment
```bash
# Apply Kubernetes manifests
kubectl apply -f k8s/

# Check deployment
kubectl get pods -l app=winkr-backend

# Check services
kubectl get services
```

#### 3. Traditional Deployment
```bash
# Build for production
make build-prod

# Build for specific platform
GOOS=linux GOARCH=amd64 go build -o bin/winkr-backend-linux cmd/api/main.go

# Deploy to server
scp bin/winkr-backend-linux user@server:/opt/winkr-backend/
ssh user@server "systemctl restart winkr-backend"
```

### Production Configuration

#### Essential Environment Variables
```bash
# Application
APP_ENV=production
APP_PORT=8080
LOG_LEVEL=info

# Database
DB_HOST=your-production-db-host
DB_PORT=5432
DB_USER=winkr_user
DB_PASSWORD=secure-password
DB_NAME=winkr_db
DB_SSL_MODE=require

# Security
JWT_SECRET=your-super-secure-jwt-secret-key
BCRYPT_COST=12
CORS_ALLOWED_ORIGINS=https://yourdomain.com

# External Services
AWS_ACCESS_KEY_ID=your-aws-access-key
AWS_SECRET_ACCESS_KEY=your-aws-secret-key
S3_BUCKET=winkr-production-storage
STRIPE_SECRET_KEY=sk_live_your-stripe-secret-key
STRIPE_WEBHOOK_SECRET=whsec_your-webhook-secret

# Monitoring
PROMETHEUS_ENABLED=true
METRICS_PORT=9091
```

### Deployment Checklist

#### Pre-Deployment
- [ ] All tests passing
- [ ] Security scan completed
- [ ] Performance tests passed
- [ ] Documentation updated
- [ ] Backup procedures verified
- [ ] Monitoring configured
- [ ] SSL certificates configured
- [ ] Database migrations tested

#### Post-Deployment
- [ ] Health checks passing
- [ ] Monitoring alerts configured
- [ ] Log collection working
- [ ] Performance metrics collected
- [ ] Security monitoring active
- [ ] Backup schedule verified

### Scaling Considerations

#### Horizontal Scaling
```bash
# Scale application instances
docker-compose -f docker/docker-compose.yml up -d --scale app=3

# Configure load balancer
# Nginx, HAProxy, or cloud load balancer
```

#### Database Scaling
```bash
# Add read replicas
# Configure connection pooling
# Implement caching strategy
# Monitor query performance
```

### Monitoring and Observability

#### Health Checks
```bash
# Basic health check
curl https://your-domain.com/health

# Detailed health check
curl https://your-domain.com/health/detailed

# Database health
curl https://your-domain.com/health/database
```

#### Metrics Collection
```bash
# Prometheus metrics
curl https://your-domain.com/metrics

# Custom metrics
curl https://your-domain.com/api/v1/admin/metrics
```

### Deployment Documentation
- [Complete Deployment Guide](docs/DEPLOYMENT.md)
- [Docker Configuration](docker/README.md)
- [Kubernetes Manifests](k8s/README.md)
- [Monitoring Setup](docs/MONITORING.md)

## 🤝 Contributing

We welcome contributions from the community! Whether you're fixing bugs, adding features, or improving documentation, your help is appreciated.

### How to Contribute

#### Quick Start
```bash
# 1. Fork the repository
# Click "Fork" on GitHub

# 2. Clone your fork
git clone https://github.com/YOUR_USERNAME/winkr-backend.git
cd winkr-backend

# 3. Create a feature branch
git checkout -b feature/your-amazing-feature

# 4. Make your changes
# Follow coding standards and write tests

# 5. Commit your changes
git commit -m "feat: add your amazing feature"

# 6. Push to your fork
git push origin feature/your-amazing-feature

# 7. Create a Pull Request
# Visit GitHub and click "New Pull Request"
```

#### Contribution Guidelines
- **Code Quality**: Follow Go conventions and project standards
- **Testing**: Write comprehensive tests for new features
- **Documentation**: Update documentation for any changes
- **Security**: Consider security implications of your changes
- **Performance**: Ensure changes don't degrade performance

#### Development Standards
```bash
# Before submitting, run quality checks
make quality

# This includes:
# - Code formatting (go fmt)
# - Linting (golangci-lint)
# - Testing (go test)
# - Security scanning (gosec)
```

### Contribution Types
- 🐛 **Bug Fixes**: Fix reported issues
- ✨ **Features**: Add new functionality
- 📚 **Documentation**: Improve documentation
- 🧪 **Tests**: Add or improve tests
- 🎨 **Code Style**: Improve code quality
- ⚡ **Performance**: Optimize performance
- 🔒 **Security**: Fix security vulnerabilities

### Recognition
- **Contributors**: Listed in README and releases
- **Hall of Fame**: Top contributors featured on website
- **Swag**: Merchandise for significant contributors
- **Maintainer Path**: Opportunity to become a maintainer

### Detailed Guidelines
- [Contributing Guide](docs/CONTRIBUTING.md)
- [Code of Conduct](docs/CODE_OF_CONDUCT.md)
- [Development Workflow](docs/DEVELOPMENT.md)

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

### License Summary
- ✅ **Commercial Use**: Allowed
- ✅ **Modification**: Allowed
- ✅ **Distribution**: Allowed
- ✅ **Private Use**: Allowed
- ⚠️ **Liability**: No warranty provided
- ⚠️ **Trademark**: No trademark rights granted

## 📞 Contact & Support

### Getting Help

#### Documentation
- 📖 [Project Overview](docs/PROJECT_OVERVIEW.md)
- 🛠 [Installation Guide](docs/INSTALLATION.md)
- 📚 [API Documentation](docs/api/)
- 🏗 [Architecture Guide](ARCHITECTURE.md)

#### Community Support
- 💬 [GitHub Discussions](https://github.com/22smeargle/winkr-backend/discussions)
- 🐛 [GitHub Issues](https://github.com/22smeargle/winkr-backend/issues)
- 📧 [Email Support](mailto:plus4822@icloud.com)

#### Reporting Issues
- 🐛 **Bug Reports**: Use GitHub issue templates
- 🔒 **Security Issues**: Email privately to [plus4822@icloud.com](mailto:plus4822@icloud.com)
- 💡 **Feature Requests**: Use GitHub Discussions
- 📖 **Documentation**: Report documentation issues

### Project Information

#### Repository
- **GitHub**: [https://github.com/22smeargle/winkr-backend](https://github.com/22smeargle/winkr-backend)
- **License**: MIT
- **Language**: Go
- **Framework**: Gin
- **Database**: PostgreSQL + Redis

#### Maintainer
- **Name**: Winkr Backend Team
- **Email**: [plus4822@icloud.com](mailto:plus4822@icloud.com)
- **Response Time**: Within 24-48 hours
- **Time Zone**: Europe/Warsaw (UTC+1)

### Additional Resources

#### Related Projects
- **Frontend**: [Winkr Mobile App](https://github.com/22smeargle/winkr-mobile) (coming soon)
- **Admin Panel**: [Winkr Admin](https://github.com/22smeargle/winkr-admin) (coming soon)

#### External Links
- **Go Documentation**: [https://golang.org/doc/](https://golang.org/doc/)
- **Gin Framework**: [https://gin-gonic.com/](https://gin-gonic.com/)
- **PostgreSQL**: [https://www.postgresql.org/docs/](https://www.postgresql.org/docs/)
- **Redis**: [https://redis.io/documentation](https://redis.io/documentation)

---

## 🙏 Acknowledgments

Thank you to all the contributors who have helped make Winkr Backend better!

### Core Contributors
- [@22smeargle](https://github.com/22smeargle) - Project creator and maintainer

### Special Thanks
- The Go community for excellent tools and libraries
- Contributors who report bugs and suggest improvements
- Users who provide valuable feedback

---

<div align="center">

**⭐ Star this repository if it helped you!**

Made with ❤️ by the Winkr Backend Team

</div>
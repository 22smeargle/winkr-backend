# Installation Guide

## Table of Contents

1. [Prerequisites and Requirements](#prerequisites-and-requirements)
2. [Step-by-Step Installation Process](#step-by-step-installation-process)
3. [Configuration Instructions](#configuration-instructions)
4. [Development Setup Guide](#development-setup-guide)
5. [Troubleshooting Section](#troubleshooting-section)
6. [Verification and Testing](#verification-and-testing)
7. [Advanced Configuration](#advanced-configuration)
8. [Performance Optimization](#performance-optimization)

## Prerequisites and Requirements

### System Requirements

#### Minimum Requirements
- **Operating System**: Linux (Ubuntu 20.04+), macOS (10.15+), Windows 10+
- **CPU**: 2 cores, 2.0 GHz
- **Memory**: 4 GB RAM
- **Storage**: 20 GB available disk space
- **Network**: Stable internet connection

#### Recommended Requirements
- **Operating System**: Linux (Ubuntu 22.04 LTS), macOS (12+), Windows 11
- **CPU**: 4 cores, 2.5 GHz
- **Memory**: 8 GB RAM
- **Storage**: 50 GB available disk space (SSD recommended)
- **Network**: High-speed internet connection

### Software Dependencies

#### Required Software
- **Go**: Version 1.21 or higher
- **Docker**: Version 20.10 or higher
- **Docker Compose**: Version 2.0 or higher
- **Git**: Version 2.30 or higher
- **Make**: Build tool (usually pre-installed)

#### Database Requirements
- **PostgreSQL**: Version 14 or higher
- **Redis**: Version 6.0 or higher

#### Development Tools (Optional but Recommended)
- **PostgreSQL Client**: For database management
- **Redis CLI**: For Redis operations
- **API Client**: Postman, Insomnia, or similar
- **IDE**: VS Code, GoLand, or similar Go-compatible IDE

### External Service Accounts

#### Required for Full Functionality
- **AWS Account**: For S3 storage and Rekognition
  - S3 bucket for file storage
  - Rekognition for photo verification
  - IAM user with appropriate permissions

- **Stripe Account**: For payment processing
  - API keys (test and production)
  - Webhook endpoint configuration

- **Email Service**: SendGrid or similar
  - API key for email sending
  - Verified sender domain

#### Optional Services
- **Monitoring**: Prometheus, Grafana
- **Logging**: ELK Stack or similar
- **CI/CD**: GitHub Actions, GitLab CI

## Step-by-Step Installation Process

### Step 1: Clone the Repository

```bash
# Clone the repository
git clone https://github.com/22smeargle/winkr-backend.git

# Navigate to the project directory
cd winkr-backend

# Verify the repository structure
ls -la
```

### Step 2: Install Go

#### Linux (Ubuntu/Debian)
```bash
# Remove any existing Go installation
sudo rm -rf /usr/local/go

# Download Go 1.21+
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz

# Extract Go to /usr/local
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz

# Add Go to PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc

# Reload bash configuration
source ~/.bashrc

# Verify Go installation
go version
```

#### macOS
```bash
# Install Go using Homebrew
brew install go@1.21

# Or download directly
wget https://go.dev/dl/go1.21.0.darwin-amd64.pkg
open go1.21.0.darwin-amd64.pkg

# Verify Go installation
go version
```

#### Windows
```bash
# Download Go installer from https://golang.org/dl/
# Run the installer and follow the setup wizard

# Verify Go installation (in PowerShell or Command Prompt)
go version
```

### Step 3: Install Docker and Docker Compose

#### Linux (Ubuntu/Debian)
```bash
# Update package index
sudo apt-get update

# Install packages to allow apt to use a repository over HTTPS
sudo apt-get install \
    ca-certificates \
    curl \
    gnupg \
    lsb-release

# Add Docker's official GPG key
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg

# Set up the stable repository
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

# Install Docker Engine
sudo apt-get update
sudo apt-get install docker-ce docker-ce-cli containerd.io docker-compose-plugin

# Add your user to the docker group
sudo usermod -aG docker $USER

# Verify Docker installation
docker --version
docker compose version
```

#### macOS
```bash
# Install Docker Desktop
brew install --cask docker

# Or download from https://www.docker.com/products/docker-desktop

# Verify Docker installation
docker --version
docker compose version
```

#### Windows
```bash
# Download Docker Desktop from https://www.docker.com/products/docker-desktop
# Run the installer and follow the setup wizard

# Verify Docker installation (in PowerShell)
docker --version
docker compose version
```

### Step 4: Set Up Environment Variables

```bash
# Copy the environment template
cp .env.example .env

# Edit the environment file
nano .env  # or use your preferred editor
```

#### Required Environment Variables
```bash
# Application Configuration
APP_ENV=development
APP_PORT=8080
APP_HOST=localhost
LOG_LEVEL=debug

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=winkr_user
DB_PASSWORD=winkr_password
DB_NAME=winkr_db
DB_SSL_MODE=disable

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
JWT_ACCESS_TOKEN_EXPIRY=15m
JWT_REFRESH_TOKEN_EXPIRY=168h

# AWS Configuration
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your-aws-access-key
AWS_SECRET_ACCESS_KEY=your-aws-secret-key
S3_BUCKET=winkr-storage

# Stripe Configuration
STRIPE_SECRET_KEY=sk_test_your-stripe-secret-key
STRIPE_WEBHOOK_SECRET=whsec_your-webhook-secret

# Email Configuration
SENDGRID_API_KEY=your-sendgrid-api-key
EMAIL_FROM=noreply@winkr.app

# Rate Limiting
RATE_LIMIT_REQUESTS_PER_MINUTE=1000
RATE_LIMIT_REQUESTS_PER_HOUR=10000
```

### Step 5: Install Dependencies

```bash
# Download Go modules
go mod download

# Verify dependencies
go mod verify

# Tidy up dependencies
go mod tidy
```

### Step 6: Start Development Environment

```bash
# Start Docker services (PostgreSQL, Redis, MinIO)
docker compose -f docker/docker-compose.dev.yml up -d

# Verify services are running
docker compose -f docker/docker-compose.dev.yml ps
```

### Step 7: Run Database Migrations

```bash
# Install migration tool (if not already installed)
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Run database migrations
make migrate-up

# Or run manually
migrate -path migrations -database "postgres://winkr_user:winkr_password@localhost:5432/winkr_db?sslmode=disable" up
```

### Step 8: Build and Run the Application

```bash
# Build the application
make build

# Or build manually
go build -o bin/winkr-backend cmd/api/main.go

# Run the application
make run

# Or run manually
./bin/winkr-backend
```

### Step 9: Verify Installation

```bash
# Check if the application is running
curl http://localhost:8080/health

# Expected response:
# {"status":"ok","timestamp":"2025-01-01T12:00:00Z"}
```

## Configuration Instructions

### Database Configuration

#### PostgreSQL Setup
```bash
# Connect to PostgreSQL
docker exec -it winkr-postgres psql -U winkr_user -d winkr_db

# Create additional databases if needed
CREATE DATABASE winkr_test;
CREATE DATABASE winkr_staging;

# Exit PostgreSQL
\q
```

#### Redis Setup
```bash
# Connect to Redis
docker exec -it winkr-redis redis-cli

# Test Redis connection
ping

# Exit Redis
exit
```

### AWS Configuration

#### S3 Bucket Setup
```bash
# Create S3 bucket (using AWS CLI)
aws s3 mb s3://winkr-storage

# Set bucket policy for public read access (if needed)
aws s3api put-bucket-policy --bucket winkr-storage --policy file://s3-policy.json
```

#### IAM User Setup
```bash
# Create IAM user with programmatic access
aws iam create-user --user-name winkr-backend

# Attach policies
aws iam attach-user-policy --user-name winkr-backend --policy-arn arn:aws:iam::aws:policy/AmazonS3FullAccess
aws iam attach-user-policy --user-name winkr-backend --policy-arn arn:aws:iam::aws:policy/AmazonRekognitionFullAccess

# Create access keys
aws iam create-access-key --user-name winkr-backend
```

### Stripe Configuration

#### Test Environment Setup
```bash
# Log in to Stripe dashboard
# Navigate to Developers > API keys
# Copy test keys to .env file

# Set up webhook endpoint
# Endpoint: https://your-domain.com/api/v1/subscriptions/webhook
# Events: customer.subscription.created, customer.subscription.updated, etc.
```

### Email Configuration

#### SendGrid Setup
```bash
# Create SendGrid account
# Verify sender domain
# Generate API key
# Add API key to .env file
```

## Development Setup Guide

### IDE Configuration

#### VS Code Setup
```bash
# Install Go extension
code --install-extension golang.go

# Install recommended extensions
code --install-extension ms-vscode.vscode-json
code --install-extension redhat.vscode-yaml
code --install-extension ms-vscode-remote.remote-containers

# Create VS Code workspace settings
mkdir -p .vscode
cat > .vscode/settings.json << EOF
{
    "go.toolsManagement.checkForUpdates": "local",
    "go.useLanguageServer": true,
    "go.gopath": "",
    "go.goroot": "",
    "go.formatTool": "goimports",
    "go.lintTool": "golangci-lint",
    "go.testFlags": ["-v"],
    "go.buildFlags": [],
    "go.testTimeout": "30s"
}
EOF
```

#### GoLand Setup
1. Open project in GoLand
2. Configure Go SDK (Settings > Go > GOROOT)
3. Set up GOPATH (Settings > Go > GOPATH)
4. Configure file watchers for go fmt and goimports
5. Set up run configurations for main.go

### Development Tools Setup

#### Makefile Commands
```bash
# View all available commands
make help

# Common development commands
make deps          # Install dependencies
make build         # Build application
make run           # Run application
make test          # Run tests
make test-cover    # Run tests with coverage
make fmt           # Format code
make lint          # Run linter
make mock          # Generate mocks
make swagger       # Generate API documentation
```

#### Database Development
```bash
# Create new migration
make migrate-create NAME=create_new_table

# Run migrations
make migrate-up

# Rollback migrations
make migrate-down

# Force migration version
make migrate-force VERSION=1

# Reset database
make migrate-reset
```

#### Testing Setup
```bash
# Run all tests
make test

# Run tests with coverage
make test-cover

# Run specific test package
go test ./internal/domain/entities/...

# Run tests with verbose output
go test -v ./...

# Run benchmarks
go test -bench=. ./...

# Run race condition tests
go test -race ./...
```

### Git Configuration

#### Git Hooks Setup
```bash
# Install pre-commit hooks
cp scripts/pre-commit .git/hooks/
chmod +x .git/hooks/pre-commit

# Install pre-push hooks
cp scripts/pre-push .git/hooks/
chmod +x .git/hooks/pre-push
```

#### Git Configuration
```bash
# Set up Git configuration
git config user.name "Your Name"
git config user.email "your.email@example.com"

# Set up Git aliases (optional)
git config alias.co checkout
git config alias.br branch
git config alias.ci commit
git config alias.st status
```

## Troubleshooting Section

### Common Installation Issues

#### Go Installation Issues
```bash
# Issue: Go command not found
# Solution: Check PATH and GOROOT
echo $PATH
which go
go env GOROOT

# Issue: Permission denied
# Solution: Fix permissions or use sudo
sudo chown -R $USER:$(id -gn $USER) /usr/local/go
```

#### Docker Issues
```bash
# Issue: Docker daemon not running
# Solution: Start Docker service
sudo systemctl start docker
sudo systemctl enable docker

# Issue: Permission denied
# Solution: Add user to docker group
sudo usermod -aG docker $USER
newgrp docker

# Issue: Port conflicts
# Solution: Check port usage and change ports
netstat -tulpn | grep :8080
# Change APP_PORT in .env file
```

#### Database Connection Issues
```bash
# Issue: Connection refused
# Solution: Check if PostgreSQL is running
docker compose -f docker/docker-compose.dev.yml ps
docker compose -f docker/docker-compose.dev.yml logs postgres

# Issue: Authentication failed
# Solution: Check database credentials
docker exec -it winkr-postgres psql -U winkr_user -d winkr_db -c "SELECT version();"

# Issue: Migration failed
# Solution: Check migration files and database state
migrate -path migrations -database "postgres://..." version
```

#### Redis Connection Issues
```bash
# Issue: Connection refused
# Solution: Check if Redis is running
docker compose -f docker/docker-compose.dev.yml ps redis
docker compose -f docker/docker-compose.dev.yml logs redis

# Issue: Authentication failed
# Solution: Check Redis password
docker exec -it winkr-redis redis-cli -a your_password ping
```

### Application Issues

#### Build Issues
```bash
# Issue: Module not found
# Solution: Clean and re-download modules
go clean -modcache
go mod download

# Issue: Compilation errors
# Solution: Check Go version and dependencies
go version
go mod tidy

# Issue: Missing dependencies
# Solution: Install missing tools
go install github.com/golangci-lint/golangci-lint/cmd/golangci-lint@latest
go install github.com/swaggo/swag/cmd/swag@latest
```

#### Runtime Issues
```bash
# Issue: Port already in use
# Solution: Kill process or change port
lsof -ti:8080 | xargs kill -9
# Or change APP_PORT in .env

# Issue: Environment variables not loaded
# Solution: Check .env file and permissions
ls -la .env
cat .env

# Issue: Database connection timeout
# Solution: Check database connectivity and configuration
docker exec -it winkr-postgres pg_isready
```

### Performance Issues

#### Slow Database Queries
```bash
# Check slow queries
docker exec -it winkr-postgres psql -U winkr_user -d winkr_db -c "
SELECT query, mean_time, calls 
FROM pg_stat_statements 
ORDER BY mean_time DESC 
LIMIT 10;"

# Check indexes
docker exec -it winkr-postgres psql -U winkr_user -d winkr_db -c "
SELECT schemaname, tablename, indexname, idx_scan 
FROM pg_stat_user_indexes 
ORDER BY idx_scan DESC;"
```

#### Memory Issues
```bash
# Check memory usage
docker stats

# Check application memory
ps aux | grep winkr-backend

# Monitor memory usage
top -p $(pgrep winkr-backend)
```

### Debugging Tips

#### Enable Debug Logging
```bash
# Set log level to debug
export LOG_LEVEL=debug

# Or update .env file
echo "LOG_LEVEL=debug" >> .env
```

#### Database Debugging
```bash
# Enable query logging
docker exec -it winkr-postgres psql -U winkr_user -d winkr_db -c "
ALTER SYSTEM SET log_statement = 'all';
SELECT pg_reload_conf();"

# Check logs
docker logs winkr-postgres
```

#### Application Debugging
```bash
# Run with race detector
go run -race cmd/api/main.go

# Run with profiler
go run -cpuprofile=cpu.prof -memprofile=mem.prof cmd/api/main.go

# Debug with Delve
dlv debug cmd/api/main.go
```

## Verification and Testing

### Health Checks

#### Application Health
```bash
# Basic health check
curl http://localhost:8080/health

# Detailed health check
curl http://localhost:8080/health/detailed

# Database health check
curl http://localhost:8080/health/database

# Redis health check
curl http://localhost:8080/health/redis
```

#### Service Health
```bash
# Check PostgreSQL
docker exec -it winkr-postgres pg_isready

# Check Redis
docker exec -it winkr-redis redis-cli ping

# Check all services
docker compose -f docker/docker-compose.dev.yml ps
```

### API Testing

#### Authentication Endpoints
```bash
# Register user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "first_name": "Test",
    "last_name": "User",
    "date_of_birth": "1990-01-01",
    "gender": "male",
    "interested_in": ["female"]
  }'

# Login user
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

#### User Management Endpoints
```bash
# Get profile (replace TOKEN with actual JWT)
curl -X GET http://localhost:8080/api/v1/users/profile \
  -H "Authorization: Bearer TOKEN"

# Update profile
curl -X PUT http://localhost:8080/api/v1/users/profile \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "bio": "Updated bio",
    "location": {
      "lat": 40.7128,
      "lng": -74.0060,
      "city": "New York",
      "country": "USA"
    }
  }'
```

### Load Testing

#### Basic Load Test
```bash
# Install Apache Bench
sudo apt-get install apache2-utils

# Run load test
ab -n 1000 -c 10 http://localhost:8080/health

# Test API endpoint
ab -n 100 -c 5 -H "Authorization: Bearer TOKEN" \
  http://localhost:8080/api/v1/users/profile
```

#### Advanced Load Testing
```bash
# Install hey (Go load testing tool)
go install github.com/rakyll/hey@latest

# Run load test
hey -n 1000 -c 10 -m GET -H "Authorization: Bearer TOKEN" \
  http://localhost:8080/api/v1/users/profile
```

## Advanced Configuration

### Production Configuration

#### Environment Variables
```bash
# Production environment
APP_ENV=production
LOG_LEVEL=info
APP_PORT=8080

# Database with SSL
DB_SSL_MODE=require
DB_SSL_CERT=/path/to/cert.pem
DB_SSL_KEY=/path/to/key.pem

# Security
BCRYPT_COST=12
JWT_ACCESS_TOKEN_EXPIRY=15m
JWT_REFRESH_TOKEN_EXPIRY=168h

# Performance
GOMAXPROCS=4
GOGC=100
```

#### Database Optimization
```bash
# PostgreSQL configuration
shared_buffers = 256MB
effective_cache_size = 1GB
maintenance_work_mem = 64MB
checkpoint_completion_target = 0.9
wal_buffers = 16MB
default_statistics_target = 100
random_page_cost = 1.1
effective_io_concurrency = 200
```

### Security Configuration

#### SSL/TLS Setup
```bash
# Generate self-signed certificate (for development)
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes

# Or use Let's Encrypt for production
sudo apt-get install certbot
sudo certbot certonly --standalone -d your-domain.com
```

#### Firewall Configuration
```bash
# Configure UFW (Ubuntu)
sudo ufw allow 22/tcp
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable

# Configure iptables (if needed)
sudo iptables -A INPUT -p tcp --dport 8080 -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 5432 -j DROP
sudo iptables -A INPUT -p tcp --dport 6379 -j DROP
```

### Monitoring Configuration

#### Prometheus Setup
```bash
# Install Prometheus
docker run -d \
  --name prometheus \
  -p 9090:9090 \
  -v $(pwd)/monitoring/prometheus.yml:/etc/prometheus/prometheus.yml \
  prom/prometheus

# Configure metrics endpoint in application
METRICS_ENABLED=true
METRICS_PORT=9091
```

#### Grafana Setup
```bash
# Install Grafana
docker run -d \
  --name grafana \
  -p 3000:3000 \
  grafana/grafana

# Configure data source and dashboards
```

## Performance Optimization

### Application Optimization

#### Go Runtime Optimization
```bash
# Set GOMAXPROCS
export GOMAXPROCS=$(nproc)

# Set GOGC for garbage collection
export GOGC=100

# Enable memory profiling
go build -ldflags="-compressdwarf=false" cmd/api/main.go
```

#### Database Optimization
```bash
# Create indexes
CREATE INDEX CONCURRENTLY idx_users_location ON users USING GIST (point(location_lng, location_lat));
CREATE INDEX CONCURRENTLY idx_users_active ON users(is_active, is_banned);
CREATE INDEX CONCURRENTLY idx_photos_user_primary ON photos(user_id, is_primary);

# Analyze tables
ANALYZE users;
ANALYZE photos;
ANALYZE matches;
```

#### Caching Strategy
```bash
# Redis configuration for caching
maxmemory 256mb
maxmemory-policy allkeys-lru

# Application-level caching
CACHE_TTL=300
CACHE_ENABLED=true
```

### Infrastructure Optimization

#### Docker Optimization
```bash
# Use multi-stage builds
# Optimize Dockerfile layers
# Use .dockerignore

# Example optimized Dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/api/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]
```

#### Load Balancer Configuration
```bash
# Nginx configuration
upstream winkr_backend {
    server 127.0.0.1:8080;
    server 127.0.0.1:8081;
    server 127.0.0.1:8082;
}

server {
    listen 80;
    server_name your-domain.com;
    
    location / {
        proxy_pass http://winkr_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

This comprehensive installation guide should help you set up Winkr Backend in various environments, from local development to production deployment. If you encounter any issues not covered in this guide, please refer to the troubleshooting section or reach out through the contact information provided in the [CONTACT.md](CONTACT.md) file.
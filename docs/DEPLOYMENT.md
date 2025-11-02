# Deployment Guide

## Table of Contents

1. [Production Deployment Steps](#production-deployment-steps)
2. [Environment Configuration](#environment-configuration)
3. [Scaling Considerations](#scaling-considerations)
4. [Monitoring Setup](#monitoring-setup)
5. [Security Considerations](#security-considerations)
6. [Deployment Strategies](#deployment-strategies)
7. [Infrastructure as Code](#infrastructure-as-code)
8. [Disaster Recovery](#disaster-recovery)
9. [Performance Optimization](#performance-optimization)
10. [Maintenance and Updates](#maintenance-and-updates)

## Production Deployment Steps

### Prerequisites

#### System Requirements
- **Operating System**: Ubuntu 20.04+ LTS or CentOS 8+
- **CPU**: Minimum 4 cores, recommended 8+ cores
- **Memory**: Minimum 8GB RAM, recommended 16GB+
- **Storage**: Minimum 100GB SSD, recommended 500GB+ SSD
- **Network**: Stable internet connection with 1Gbps+ bandwidth

#### Software Requirements
- **Docker**: Version 20.10+ with Docker Compose 2.0+
- **Kubernetes**: Version 1.24+ (for K8s deployment)
- **PostgreSQL**: Version 14+ (if not using container)
- **Redis**: Version 6+ (if not using container)
- **Nginx**: Version 1.20+ (for load balancing)
- **SSL Certificate**: Valid SSL certificate for HTTPS

#### External Services
- **AWS Account**: S3, Rekognition, and Route53 access
- **Stripe Account**: Production API keys and webhooks
- **Email Service**: SendGrid or similar with verified domain
- **Monitoring**: Prometheus, Grafana, or similar
- **Logging**: ELK Stack or similar centralized logging

### Step 1: Environment Preparation

#### Server Setup
```bash
# Update system packages
sudo apt update && sudo apt upgrade -y

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Install Nginx
sudo apt install nginx -y

# Install SSL certificate (Let's Encrypt)
sudo apt install certbot python3-certbot-nginx -y
sudo certbot --nginx -d yourdomain.com
```

#### Firewall Configuration
```bash
# Configure UFW firewall
sudo ufw enable
sudo ufw allow ssh
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw allow from 10.0.0.0/8 to any port 5432  # Database access
sudo ufw allow from 10.0.0.0/8 to any port 6379  # Redis access
```

### Step 2: Application Deployment

#### Docker Deployment (Recommended)
```bash
# Clone repository
git clone https://github.com/22smeargle/winkr-backend.git
cd winkr-backend

# Configure production environment
cp .env.example .env.production
nano .env.production

# Build production image
docker build -f docker/Dockerfile -t winkr-backend:latest .

# Create production docker-compose.yml
cat > docker-compose.prod.yml << EOF
version: '3.8'

services:
  app:
    image: winkr-backend:latest
    ports:
      - "8080:8080"
    environment:
      - APP_ENV=production
      - DB_HOST=postgres
      - REDIS_HOST=redis
    depends_on:
      - postgres
      - redis
    restart: unless-stopped
    deploy:
      replicas: 3
      resources:
        limits:
          cpus: '1.0'
          memory: 1G
        reservations:
          cpus: '0.5'
          memory: 512M

  postgres:
    image: postgres:14-alpine
    environment:
      POSTGRES_DB: ${DB_NAME}
      POSTGRES_USER: ${DB_USER}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./backups:/backups
    restart: unless-stopped
    deploy:
      resources:
        limits:
          cpus: '2.0'
          memory: 2G

  redis:
    image: redis:6-alpine
    command: redis-server --appendonly yes --requirepass ${REDIS_PASSWORD}
    volumes:
      - redis_data:/data
    restart: unless-stopped
    deploy:
      resources:
        limits:
          cpus: '1.0'
          memory: 1G

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf
      - ./ssl:/etc/ssl/certs
    depends_on:
      - app
    restart: unless-stopped

volumes:
  postgres_data:
  redis_data:
EOF

# Deploy application
docker-compose -f docker-compose.prod.yml --env-file .env.production up -d
```

#### Kubernetes Deployment
```bash
# Create namespace
kubectl create namespace winkr

# Create ConfigMap
kubectl create configmap winkr-config \
  --from-env-file=.env.production \
  --namespace=winkr

# Create secrets
kubectl create secret generic winkr-secrets \
  --from-literal=db-password=${DB_PASSWORD} \
  --from-literal=jwt-secret=${JWT_SECRET} \
  --from-literal=stripe-secret=${STRIPE_SECRET_KEY} \
  --namespace=winkr

# Apply manifests
kubectl apply -f k8s/ --namespace=winkr

# Check deployment status
kubectl get pods --namespace=winkr
kubectl get services --namespace=winkr
```

### Step 3: Load Balancer Configuration

#### Nginx Configuration
```nginx
# /etc/nginx/sites-available/winkr-backend
upstream winkr_backend {
    least_conn;
    server 127.0.0.1:8080 max_fails=3 fail_timeout=30s;
    server 127.0.0.1:8081 max_fails=3 fail_timeout=30s;
    server 127.0.0.1:8082 max_fails=3 fail_timeout=30s;
}

server {
    listen 80;
    server_name yourdomain.com www.yourdomain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name yourdomain.com www.yourdomain.com;

    ssl_certificate /etc/ssl/certs/yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/ssl/certs/yourdomain.com/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512;
    ssl_prefer_server_ciphers off;

    # Security headers
    add_header X-Frame-Options DENY;
    add_header X-Content-Type-Options nosniff;
    add_header X-XSS-Protection "1; mode=block";
    add_header Strict-Transport-Security "max-age=63072000; includeSubDomains; preload";

    # Rate limiting
    limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;
    limit_req zone=api burst=20 nodelay;

    location / {
        proxy_pass http://winkr_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # WebSocket support
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        
        # Timeouts
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    # Health check endpoint
    location /health {
        proxy_pass http://winkr_backend;
        access_log off;
    }

    # Static files (if any)
    location /static/ {
        alias /var/www/winkr-backend/static/;
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
}
```

### Step 4: Database Setup

#### PostgreSQL Production Configuration
```bash
# /etc/postgresql/14/main/postgresql.conf
# Memory settings
shared_buffers = 256MB
effective_cache_size = 1GB
work_mem = 4MB
maintenance_work_mem = 64MB

# Connection settings
max_connections = 200
listen_addresses = '*'

# Performance settings
checkpoint_completion_target = 0.9
wal_buffers = 16MB
default_statistics_target = 100
random_page_cost = 1.1
effective_io_concurrency = 200

# Logging
log_min_duration_statement = 1000
log_checkpoints = on
log_connections = on
log_disconnections = on
```

#### Database Security
```bash
# /etc/postgresql/14/main/pg_hba.conf
# Local connections
local   all             postgres                                peer
local   all             all                                     md5

# Application connections
host     all             all             10.0.0.0/8           md5
host     all             all             127.0.0.1/32            md5

# SSL connections
hostssl  all             all             0.0.0.0/0               md5
```

### Step 5: Redis Configuration

#### Redis Production Configuration
```bash
# /etc/redis/redis.conf
# Memory
maxmemory 512mb
maxmemory-policy allkeys-lru

# Persistence
save 900 1
save 300 10
save 60 10000
appendonly yes
appendfsync everysec

# Security
requirepass your-redis-password
rename-command FLUSH ""
rename-command CONFIG ""
rename-command DEBUG ""

# Network
bind 127.0.0.1 10.0.0.1
port 6379
timeout 300

# Performance
tcp-keepalive 300
tcp-backlog 511
```

## Environment Configuration

### Production Environment Variables

#### Essential Configuration
```bash
# Application
APP_ENV=production
APP_PORT=8080
APP_HOST=0.0.0.0
LOG_LEVEL=info

# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=winkr_user
DB_PASSWORD=secure-password-change-this
DB_NAME=winkr_production
DB_SSL_MODE=require
DB_MAX_CONNECTIONS=50
DB_MAX_IDLE_CONNECTIONS=10

# Redis
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=secure-redis-password
REDIS_DB=0
REDIS_MAX_CONNECTIONS=20

# Security
JWT_SECRET=your-super-secure-jwt-secret-key-change-this
JWT_ACCESS_TOKEN_EXPIRY=15m
JWT_REFRESH_TOKEN_EXPIRY=168h
BCRYPT_COST=12

# CORS
CORS_ALLOWED_ORIGINS=https://yourdomain.com,https://www.yourdomain.com
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=Content-Type,Authorization

# Rate Limiting
RATE_LIMIT_REQUESTS_PER_MINUTE=100
RATE_LIMIT_REQUESTS_PER_HOUR=1000
RATE_LIMIT_BURST=200

# File Upload
MAX_FILE_SIZE=10485760  # 10MB
ALLOWED_FILE_TYPES=jpg,jpeg,png,gif
PHOTO_QUALITY=85

# AWS Services
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your-aws-access-key
AWS_SECRET_ACCESS_KEY=your-aws-secret-key
S3_BUCKET=winkr-production-storage
S3_REGION=us-east-1

# Stripe
STRIPE_SECRET_KEY=sk_live_your-stripe-secret-key
STRIPE_WEBHOOK_SECRET=whsec_your-webhook-secret
STRIPE_SUCCESS_URL=https://yourdomain.com/success
STRIPE_CANCEL_URL=https://yourdomain.com/cancel

# Email
SENDGRID_API_KEY=your-sendgrid-api-key
EMAIL_FROM=noreply@yourdomain.com
EMAIL_REPLY_TO=support@yourdomain.com

# Monitoring
PROMETHEUS_ENABLED=true
METRICS_PORT=9091
HEALTH_CHECK_ENABLED=true
HEALTH_CHECK_INTERVAL=30s
```

#### Advanced Configuration
```bash
# Performance
GOMAXPROCS=4
GOGC=100
GOTRACEBACK=1

# Logging
LOG_FORMAT=json
LOG_OUTPUT=file
LOG_FILE=/var/log/winkr-backend/app.log
LOG_MAX_SIZE=100MB
LOG_MAX_BACKUPS=10
LOG_MAX_AGE=30

# Caching
CACHE_TTL=300
CACHE_CLEANUP_INTERVAL=3600
ENABLE_QUERY_CACHE=true

# Background Jobs
BACKGROUND_WORKERS=10
JOB_QUEUE_SIZE=1000
JOB_RETRY_ATTEMPTS=3
JOB_RETRY_DELAY=60s

# WebSocket
WEBSOCKET_PING_INTERVAL=30s
WEBSOCKET_MAX_CONNECTIONS=10000
WEBSOCKET_MESSAGE_SIZE_LIMIT=16384
```

### Configuration Management

#### Environment-Specific Configs
```bash
# Directory structure
/configs/
├── production.env
├── staging.env
├── development.env
└── testing.env

/secrets/
├── production-secrets.env
├── staging-secrets.env
└── development-secrets.env
```

#### Configuration Validation
```bash
# Validate configuration before deployment
./scripts/validate-config.sh

# Check required environment variables
./scripts/check-env.sh

# Test database connectivity
./scripts/test-db-connection.sh

# Test external services
./scripts/test-external-services.sh
```

## Scaling Considerations

### Horizontal Scaling

#### Application Scaling
```bash
# Docker Compose scaling
docker-compose -f docker-compose.prod.yml up -d --scale app=5

# Kubernetes Horizontal Pod Autoscaler
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: winkr-backend-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: winkr-backend
  minReplicas: 3
  maxReplicas: 20
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

#### Database Scaling
```bash
# Read replicas configuration
# postgresql.conf
wal_level = replica
max_wal_senders = 3
max_replication_slots = 3

# Create read replicas
docker run -d --name postgres-replica1 \
  -e POSTGRES_MASTER_SERVICE=postgres \
  -e POSTGRES_REPLICATION_USER=replicator \
  -e POSTGRES_REPLICATION_PASSWORD=replicator-password \
  postgres:14-alpine

# Connection pooling
pgbouncer -d /etc/pgbouncer.ini
```

#### Redis Scaling
```bash
# Redis Cluster configuration
redis-cli --cluster create \
  127.0.0.1:7000 127.0.0.1:7001 127.0.0.1:7002 \
  127.0.0.1:7003 127.0.0.1:7004 127.0.0.1:7005 \
  --cluster-replicas 1

# Redis Sentinel for high availability
port 26379
sentinel monitor mymaster 127.0.0.1 6379 2
sentinel down-after-milliseconds mymaster 5000
sentinel failover-timeout mymaster 10000
```

### Vertical Scaling

#### Resource Optimization
```bash
# Monitor resource usage
docker stats
kubectl top pods

# Optimize resource limits
resources:
  requests:
    cpu: 500m
    memory: 512Mi
  limits:
    cpu: 2000m
    memory: 2Gi
```

#### Performance Tuning
```bash
# Go runtime optimization
export GOMAXPROCS=$(nproc)
export GOGC=100

# Database optimization
shared_buffers = 25% of RAM
effective_cache_size = 75% of RAM
work_mem = (RAM - shared_buffers) / max_connections

# Redis optimization
maxmemory = 80% of available RAM
maxmemory-policy = allkeys-lru
```

## Monitoring Setup

### Prometheus Configuration

#### Prometheus Server
```yaml
# prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "winkr-rules.yml"

scrape_configs:
  - job_name: 'winkr-backend'
    static_configs:
      - targets: ['app:9091']
    metrics_path: /metrics
    scrape_interval: 10s

  - job_name: 'postgres'
    static_configs:
      - targets: ['postgres-exporter:9187']

  - job_name: 'redis'
    static_configs:
      - targets: ['redis-exporter:9121']

  - job_name: 'nginx'
    static_configs:
      - targets: ['nginx-exporter:9113']

alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - alertmanager:9093
```

#### Alerting Rules
```yaml
# winkr-rules.yml
groups:
  - name: winkr-backend
    rules:
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.1
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ $value }} errors per second"

      - alert: HighMemoryUsage
        expr: (node_memory_MemTotal_bytes - node_memory_MemAvailable_bytes) / node_memory_MemTotal_bytes > 0.9
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High memory usage"
          description: "Memory usage is above 90%"

      - alert: DatabaseDown
        expr: up{job="postgres"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Database is down"
          description: "PostgreSQL database is not responding"
```

### Grafana Dashboards

#### Application Dashboard
```json
{
  "dashboard": {
    "title": "Winkr Backend Overview",
    "panels": [
      {
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total[5m])",
            "legendFormat": "{{method}} {{status}}"
          }
        ]
      },
      {
        "title": "Response Time",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "95th percentile"
          }
        ]
      },
      {
        "title": "Error Rate",
        "type": "singlestat",
        "targets": [
          {
            "expr": "rate(http_requests_total{status=~\"5..\"}[5m])",
            "legendFormat": "Error Rate"
          }
        ]
      }
    ]
  }
}
```

### Log Aggregation

#### ELK Stack Setup
```yaml
# docker-compose.logging.yml
version: '3.8'

services:
  elasticsearch:
    image: docker.elastic.co/elasticsearch/elasticsearch:7.15.0
    environment:
      - discovery.type=single-node
      - "ES_JAVA_OPTS=-Xms512m -Xmx512m"
    volumes:
      - elasticsearch_data:/usr/share/elasticsearch/data
    ports:
      - "9200:9200"

  logstash:
    image: docker.elastic.co/logstash/logstash:7.15.0
    volumes:
      - ./logstash/pipeline:/usr/share/logstash/pipeline
      - ./logstash/config:/usr/share/logstash/config
    ports:
      - "5044:5044"
    depends_on:
      - elasticsearch

  kibana:
    image: docker.elastic.co/kibana/kibana:7.15.0
    ports:
      - "5601:5601"
    environment:
      ELASTICSEARCH_HOSTS: http://elasticsearch:9200
    depends_on:
      - elasticsearch

volumes:
  elasticsearch_data:
```

## Security Considerations

### Network Security

#### Firewall Configuration
```bash
# UFW configuration
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow ssh
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# Fail2Ban configuration
sudo apt install fail2ban -y
sudo systemctl enable fail2ban
sudo systemctl start fail2ban

# /etc/fail2ban/jail.local
[sshd]
enabled = true
port = ssh
filter = sshd
logpath = /var/log/auth.log
maxretry = 3
bantime = 3600
findtime = 600
```

#### SSL/TLS Configuration
```bash
# Generate strong SSL certificate
sudo certbot --nginx -d yourdomain.com --rsa-key-size 4096

# Configure SSL hardening
# /etc/nginx/snippets/ssl-params.conf
ssl_protocols TLSv1.2 TLSv1.3;
ssl_ciphers ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384;
ssl_prefer_server_ciphers off;
ssl_session_cache shared:SSL:10m;
ssl_session_timeout 10m;
ssl_stapling on;
ssl_stapling_verify on;
```

### Application Security

#### Security Headers
```nginx
# Security headers in Nginx
add_header X-Frame-Options DENY;
add_header X-Content-Type-Options nosniff;
add_header X-XSS-Protection "1; mode=block";
add_header Strict-Transport-Security "max-age=63072000; includeSubDomains; preload";
add_header Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'";
add_header Referrer-Policy "strict-origin-when-cross-origin";
```

#### Input Validation
```go
// Security middleware
func SecurityMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Rate limiting
        if !rateLimiter.Allow(c.ClientIP()) {
            c.JSON(429, gin.H{"error": "Rate limit exceeded"})
            c.Abort()
            return
        }

        // Input validation
        if c.Request.Method == "POST" || c.Request.Method == "PUT" {
            if !validateInput(c) {
                c.JSON(400, gin.H{"error": "Invalid input"})
                c.Abort()
                return
            }
        }

        // Security headers
        c.Header("X-Content-Type-Options", "nosniff")
        c.Header("X-Frame-Options", "DENY")
        c.Header("X-XSS-Protection", "1; mode=block")

        c.Next()
    }
}
```

### Database Security

#### PostgreSQL Security
```bash
# Enable SSL
ssl = on
ssl_cert_file = '/etc/ssl/certs/server.crt'
ssl_key_file = '/etc/ssl/certs/server.key'

# Row-level security
CREATE POLICY user_isolation_policy ON users
    FOR ALL
    TO app_user
    USING (id = current_setting('app.user_id'))
    WITH CHECK (true);

ALTER TABLE users ENABLE ROW LEVEL SECURITY;
```

#### Redis Security
```bash
# Redis security configuration
requirepass your-strong-password
rename-command FLUSH ""
rename-command CONFIG ""
rename-command DEBUG ""
rename-command EVAL ""

# Network security
bind 127.0.0.1
protected-mode yes
```

## Deployment Strategies

### Blue-Green Deployment

#### Deployment Script
```bash
#!/bin/bash
# blue-green-deploy.sh

CURRENT_ENV=$(docker service ls --filter name=winkr-backend --format "{{.Name}}" | grep -E "(blue|green)" | head -1)
NEW_ENV="green"

if [[ $CURRENT_ENV == "green" ]]; then
    NEW_ENV="blue"
fi

echo "Deploying to $NEW_ENV environment"

# Deploy new version
docker service create \
    --name winkr-backend-$NEW_ENV \
    --replicas 3 \
    --network winkr-network \
    --env-file .env.production \
    winkr-backend:latest

# Health check
echo "Waiting for health check..."
sleep 30

HEALTH_CHECK=$(curl -s http://winkr-backend-$NEW_ENV:8080/health | jq -r '.status')
if [[ $HEALTH_CHECK != "ok" ]]; then
    echo "Health check failed, rolling back..."
    docker service rm winkr-backend-$NEW_ENV
    exit 1
fi

# Switch traffic
echo "Switching traffic to $NEW_ENV"
docker service update \
    --label traefik.http.routers.winkr-backend.rule=Host(\`yourdomain.com\`) \
    winkr-backend-$NEW_ENV

# Clean up old environment
echo "Cleaning up $CURRENT_ENV environment"
docker service rm winkr-backend-$CURRENT_ENV

echo "Deployment completed successfully"
```

### Canary Deployment

#### Canary Configuration
```yaml
# kubernetes/canary-deployment.yaml
apiVersion: argoproj.io/v1alpha1
kind: Rollout
metadata:
  name: winkr-backend-canary
spec:
  replicas: 10
  strategy:
    canary:
      steps:
      - setWeight: 10
      - pause: {duration: 10m}
      - setWeight: 25
      - pause: {duration: 10m}
      - setWeight: 50
      - pause: {duration: 10m}
      - setWeight: 100
      canaryService: winkr-backend-canary
      stableService: winkr-backend-stable
  selector:
    matchLabels:
      app: winkr-backend
  template:
    metadata:
      labels:
        app: winkr-backend
    spec:
      containers:
      - name: winkr-backend
        image: winkr-backend:latest
        ports:
        - containerPort: 8080
```

### Rolling Deployment

#### Rolling Update Strategy
```yaml
# kubernetes/rolling-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: winkr-backend
spec:
  replicas: 5
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
      maxSurge: 2
  selector:
    matchLabels:
      app: winkr-backend
  template:
    metadata:
      labels:
        app: winkr-backend
    spec:
      containers:
      - name: winkr-backend
        image: winkr-backend:latest
        ports:
        - containerPort: 8080
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 5
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
```

## Infrastructure as Code

### Terraform Configuration

#### AWS Infrastructure
```hcl
# main.tf
provider "aws" {
  region = var.aws_region
}

# VPC Configuration
resource "aws_vpc" "main" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = "winkr-vpc"
  }
}

# ECS Cluster
resource "aws_ecs_cluster" "main" {
  name = "winkr-cluster"

  setting {
    name  = "containerInsights"
    value = "enabled"
  }
}

# RDS Database
resource "aws_db_instance" "postgres" {
  identifier = "winkr-postgres"
  engine     = "postgres"
  instance_class = "db.m5.large"
  allocated_storage     = 100
  storage_type        = "gp2"
  engine_version      = "14.6"
  parameter_group_name = "default.postgres14"

  username = var.db_username
  password = var.db_password

  skip_final_snapshot = false
  final_snapshot_identifier = "winkr-postgres-final-snapshot"

  tags = {
    Name = "winkr-postgres"
  }
}

# ElastiCache Redis
resource "aws_elasticache_subnet_group" "main" {
  name       = "winkr-cache-subnet"
  subnet_ids = aws_subnet.private[*].id
}

resource "aws_elasticache_cluster" "redis" {
  cluster_id           = "winkr-redis"
  engine               = "redis"
  node_type            = "cache.m5.large"
  num_cache_nodes      = 3
  parameter_group_name = "default.redis6.x"
  port                 = 6379
  subnet_group_name    = aws_elasticache_subnet_group.main.name
  security_group_ids   = [aws_security_group.redis.id]

  tags = {
    Name = "winkr-redis"
  }
}
```

#### Kubernetes Manifests
```yaml
# k8s/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: winkr
  labels:
    name: winkr

---
# k8s/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: winkr-config
  namespace: winkr
data:
  APP_ENV: "production"
  LOG_LEVEL: "info"
  DB_HOST: "postgres-service"
  REDIS_HOST: "redis-service"

---
# k8s/secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: winkr-secrets
  namespace: winkr
type: Opaque
data:
  db-password: <base64-encoded-password>
  jwt-secret: <base64-encoded-secret>
  stripe-secret: <base64-encoded-stripe-key>
```

### Ansible Playbooks

#### Server Provisioning
```yaml
# playbook.yml
---
- hosts: winkr-servers
  become: yes
  vars:
    app_version: "latest"
    domain: "yourdomain.com"
  
  tasks:
    - name: Update system packages
      apt:
        update_cache: yes
        upgrade: dist

    - name: Install Docker
      apt:
        name: docker.io
        state: present

    - name: Install Docker Compose
      get_url:
        url: "https://github.com/docker/compose/releases/latest/download/docker-compose-{{ ansible_system }}-{{ ansible_architecture }}"
        dest: /usr/local/bin/docker-compose
        mode: '0755'

    - name: Create application directory
      file:
        path: /opt/winkr-backend
        state: directory
        mode: '0755'

    - name: Deploy application
      docker_compose:
        project_src: /opt/winkr-backend
        files:
          - docker-compose.prod.yml
        state: present
```

## Disaster Recovery

### Backup Strategy

#### Database Backups
```bash
#!/bin/bash
# backup-database.sh

BACKUP_DIR="/backups/postgres"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="winkr_backup_$DATE.sql"

# Create backup directory
mkdir -p $BACKUP_DIR

# Create database backup
pg_dump -h $DB_HOST -U $DB_USER -d $DB_NAME \
  --no-password --verbose --format=custom \
  --file=$BACKUP_DIR/$BACKUP_FILE

# Compress backup
gzip $BACKUP_DIR/$BACKUP_FILE

# Upload to S3
aws s3 cp $BACKUP_DIR/$BACKUP_FILE.gz \
  s3://winkr-backups/database/

# Clean up old backups (keep last 30 days)
find $BACKUP_DIR -name "*.sql.gz" -mtime +30 -delete

echo "Database backup completed: $BACKUP_FILE.gz"
```

#### File Backups
```bash
#!/bin/bash
# backup-files.sh

BACKUP_DIR="/backups/files"
DATE=$(date +%Y%m%d_%H%M%S)

# Sync S3 to local backup
aws s3 sync s3://winkr-production-storage/ \
  $BACKUP_DIR/$DATE/ --delete

# Create compressed archive
tar -czf $BACKUP_DIR/files_backup_$DATE.tar.gz \
  $BACKUP_DIR/$DATE/

# Upload to backup S3 bucket
aws s3 cp $BACKUP_DIR/files_backup_$DATE.tar.gz \
  s3://winkr-backups/files/

# Clean up
rm -rf $BACKUP_DIR/$DATE

echo "File backup completed: files_backup_$DATE.tar.gz"
```

### Recovery Procedures

#### Database Recovery
```bash
#!/bin/bash
# restore-database.sh

BACKUP_FILE=$1

if [[ -z $BACKUP_FILE ]]; then
    echo "Usage: $0 <backup_file>"
    exit 1
fi

# Download backup from S3
aws s3 cp s3://winkr-backups/database/$BACKUP_FILE.gz ./

# Extract backup
gunzip $BACKUP_FILE.gz

# Stop application
docker-compose -f docker-compose.prod.yml down

# Restore database
pg_restore -h $DB_HOST -U $DB_USER -d $DB_NAME \
  --no-password --verbose --clean --if-exists \
  $BACKUP_FILE

# Start application
docker-compose -f docker-compose.prod.yml up -d

echo "Database restore completed"
```

#### Application Recovery
```bash
#!/bin/bash
# disaster-recovery.sh

# Check system health
if ! curl -f http://localhost:8080/health; then
    echo "Application is down, initiating recovery..."
    
    # Check database
    if ! pg_isready -h $DB_HOST -U $DB_USER; then
        echo "Database is down, restoring from backup..."
        ./scripts/restore-database.sh latest_backup.sql.gz
    fi
    
    # Check Redis
    if ! redis-cli -h $REDIS_HOST ping; then
        echo "Redis is down, restarting..."
        docker-compose restart redis
    fi
    
    # Restart application
    docker-compose restart app
    
    # Verify recovery
    sleep 30
    if curl -f http://localhost:8080/health; then
        echo "Recovery completed successfully"
    else
        echo "Recovery failed, manual intervention required"
        exit 1
    fi
fi
```

## Performance Optimization

### Database Optimization

#### Query Optimization
```sql
-- Create optimal indexes
CREATE INDEX CONCURRENTLY idx_users_location ON users USING GIST (point(location_lng, location_lat));
CREATE INDEX CONCURRENTLY idx_users_active ON users(is_active, is_banned, last_active DESC);
CREATE INDEX CONCURRENTLY idx_photos_user_primary ON photos(user_id, is_primary);
CREATE INDEX CONCURRENTLY idx_messages_conversation_time ON messages(conversation_id, created_at DESC);

-- Analyze tables for query planner
ANALYZE users;
ANALYZE photos;
ANALYZE messages;
ANALYZE matches;

-- Monitor slow queries
SELECT query, mean_time, calls, total_time
FROM pg_stat_statements
ORDER BY mean_time DESC
LIMIT 10;
```

#### Connection Pooling
```go
// Database connection pool configuration
func NewDatabase(config *Config) (*sql.DB, error) {
    dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
        config.DBHost, config.DBPort, config.DBUser, config.DBPassword, config.DBName, config.DBSSLMode)
    
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        return nil, err
    }
    
    // Configure connection pool
    db.SetMaxOpenConns(config.DBMaxConnections)
    db.SetMaxIdleConns(config.DBMaxIdleConnections)
    db.SetConnMaxLifetime(time.Hour)
    db.SetConnMaxIdleTime(30 * time.Minute)
    
    return db, nil
}
```

### Caching Strategy

#### Redis Caching
```go
// Multi-level caching
type CacheService struct {
    redis    *redis.Client
    local    *sync.Map
    ttl      time.Duration
}

func (c *CacheService) Get(key string) (interface{}, error) {
    // Check local cache first
    if value, ok := c.local.Load(key); ok {
        return value, nil
    }
    
    // Check Redis cache
    value, err := c.redis.Get(key).Result()
    if err == nil {
        // Cache in local for faster access
        c.local.Store(key, value)
        return value, nil
    }
    
    return nil, err
}

func (c *CacheService) Set(key string, value interface{}, ttl time.Duration) error {
    // Set in both local and Redis
    c.local.Store(key, value)
    return c.redis.Set(key, value, ttl).Err()
}
```

### Application Optimization

#### Go Runtime Optimization
```go
// Runtime optimization in main.go
func main() {
    // Set GOMAXPROCS
    runtime.GOMAXPROCS(runtime.NumCPU())
    
    // Set GOGC for garbage collection
    debug.SetGCPercent(100)
    
    // Enable profiling
    if os.Getenv("APP_ENV") == "production" {
        go func() {
            log.Println(http.ListenAndServe("localhost:6060", nil))
        }()
    }
    
    // Start application
    app := setupApplication()
    app.Run(":8080")
}
```

#### Memory Optimization
```go
// Object pooling for memory efficiency
var bufferPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 0, 1024)
    },
}

func processRequest(data []byte) {
    buffer := bufferPool.Get().([]byte)
    defer bufferPool.Put(buffer[:0])
    
    // Use buffer for processing
    buffer = append(buffer, data...)
    
    // Process buffer
    result := processData(buffer)
    
    return result
}
```

## Maintenance and Updates

### Update Procedures

#### Rolling Updates
```bash
#!/bin/bash
# rolling-update.sh

NEW_VERSION=$1
if [[ -z $NEW_VERSION ]]; then
    echo "Usage: $0 <new_version>"
    exit 1
fi

echo "Starting rolling update to version $NEW_VERSION"

# Pull new version
docker pull winkr-backend:$NEW_VERSION

# Update one replica at a time
for i in {1..3}; do
    echo "Updating replica $i/3"
    
    # Stop one replica
    docker-compose -f docker-compose.prod.yml up -d --scale app=2
    
    # Wait for health check
    sleep 30
    
    # Update with new version
    docker-compose -f docker-compose.prod.yml up -d --scale app=3
    
    # Verify health
    if curl -f http://localhost:8080/health; then
        echo "Replica $i updated successfully"
    else
        echo "Health check failed, rolling back..."
        docker-compose -f docker-compose.prod.yml up -d --scale app=3
        exit 1
    fi
done

echo "Rolling update completed successfully"
```

#### Database Migrations
```bash
#!/bin/bash
# migrate-database.sh

NEW_VERSION=$1

echo "Running database migrations to version $NEW_VERSION"

# Backup current database
./scripts/backup-database.sh

# Run migrations
migrate -path migrations \
  -database "postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=$DB_SSL_MODE" \
  up

# Verify migration
migrate -path migrations \
  -database "postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=$DB_SSL_MODE" \
  version

echo "Database migration completed"
```

### Health Monitoring

#### Health Check Script
```bash
#!/bin/bash
# health-check.sh

HEALTH_URL="http://localhost:8080/health"
ALERT_EMAIL="admin@yourdomain.com"

# Check application health
RESPONSE=$(curl -s -w "%{http_code}" $HEALTH_URL)
HTTP_CODE="${RESPONSE: -3}"

if [[ $HTTP_CODE != "200" ]]; then
    echo "Health check failed with HTTP code: $HTTP_CODE"
    
    # Send alert
    echo "Winkr Backend health check failed" | \
      mail -s "Health Check Alert" $ALERT_EMAIL
    
    # Attempt recovery
    ./scripts/disaster-recovery.sh
    
    exit 1
fi

# Check database health
if ! pg_isready -h $DB_HOST -U $DB_USER; then
    echo "Database health check failed"
    echo "Database health check failed" | \
      mail -s "Database Health Alert" $ALERT_EMAIL
    exit 1
fi

# Check Redis health
if ! redis-cli -h $REDIS_HOST ping; then
    echo "Redis health check failed"
    echo "Redis health check failed" | \
      mail -s "Redis Health Alert" $ALERT_EMAIL
    exit 1
fi

echo "All health checks passed"
exit 0
```

This comprehensive deployment guide provides everything needed to deploy Winkr Backend in production environments, from basic setup to advanced scaling and monitoring configurations. For additional support, please refer to the [contact information](CONTACT.md) or [installation guide](INSTALLATION.md).
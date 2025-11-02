# Dating Application Backend Architecture

## Table of Contents

1. [Project Structure](#project-structure)
2. [Database Schema Design](#database-schema-design)
3. [API Endpoint Specifications](#api-endpoint-specifications)
4. [Security Architecture](#security-architecture)
5. [External Service Integrations](#external-service-integrations)
6. [Technology Stack](#technology-stack)
7. [Deployment Architecture](#deployment-architecture)

## Project Structure

The backend follows Clean Architecture principles with clear separation of concerns:

```
backend/
├── cmd/
│   └── api/
│       └── main.go                 # Application entry point
├── internal/
│   ├── domain/                     # Business logic and entities
│   │   ├── entities/               # Core business entities
│   │   │   ├── user.go
│   │   │   ├── photo.go
│   │   │   ├── message.go
│   │   │   ├── match.go
│   │   │   ├── report.go
│   │   │   └── subscription.go
│   │   ├── repositories/          # Repository interfaces
│   │   │   ├── user_repository.go
│   │   │   ├── photo_repository.go
│   │   │   ├── message_repository.go
│   │   │   ├── match_repository.go
│   │   │   ├── report_repository.go
│   │   │   └── subscription_repository.go
│   │   ├── services/               # Business logic services
│   │   │   ├── auth_service.go
│   │   │   ├── user_service.go
│   │   │   ├── photo_service.go
│   │   │   ├── message_service.go
│   │   │   ├── match_service.go
│   │   │   ├── report_service.go
│   │   │   └── subscription_service.go
│   │   └── valueobjects/           # Value objects and enums
│   │       ├── gender.go
│   │       ├── relationship_status.go
│   │       └── verification_status.go
│   ├── application/                # Application layer
│   │   ├── usecases/               # Use cases
│   │   │   ├── auth/
│   │   │   │   ├── register_usecase.go
│   │   │   │   ├── login_usecase.go
│   │   │   │   └── refresh_token_usecase.go
│   │   │   ├── user/
│   │   │   │   ├── get_profile_usecase.go
│   │   │   │   ├── update_profile_usecase.go
│   │   │   │   └── delete_account_usecase.go
│   │   │   ├── photo/
│   │   │   │   ├── upload_photo_usecase.go
│   │   │   │   ├── delete_photo_usecase.go
│   │   │   │   └── verify_photo_usecase.go
│   │   │   ├── matching/
│   │   │   │   ├── swipe_usecase.go
│   │   │   │   ├── get_matches_usecase.go
│   │   │   │   └── get_potential_matches_usecase.go
│   │   │   ├── messaging/
│   │   │   │   ├── send_message_usecase.go
│   │   │   │   ├── get_conversations_usecase.go
│   │   │   │   └── get_messages_usecase.go
│   │   │   ├── subscription/
│   │   │   │   ├── create_subscription_usecase.go
│   │   │   │   ├── cancel_subscription_usecase.go
│   │   │   │   └── webhook_handler_usecase.go
│   │   │   └── reporting/
│   │   │       ├── create_report_usecase.go
│   │   │       └── review_report_usecase.go
│   │   └── dto/                    # Data transfer objects
│   │       ├── auth_dto.go
│   │       ├── user_dto.go
│   │       ├── photo_dto.go
│   │       ├── message_dto.go
│   │       ├── match_dto.go
│   │       ├── report_dto.go
│   │       └── subscription_dto.go
│   ├── infrastructure/             # Infrastructure layer
│   │   ├── database/               # Database implementations
│   │   │   ├── postgres/
│   │   │   │   ├── connection.go
│   │   │   │   ├── migrations/
│   │   │   │   └── repositories/
│   │   │   │       ├── user_repository_impl.go
│   │   │   │       ├── photo_repository_impl.go
│   │   │   │       ├── message_repository_impl.go
│   │   │   │       ├── match_repository_impl.go
│   │   │   │       ├── report_repository_impl.go
│   │   │   │       └── subscription_repository_impl.go
│   │   │   └── redis/
│   │   │       ├── connection.go
│   │   │       └── cache_repository.go
│   │   ├── storage/                # File storage
│   │   │   ├── s3/
│   │   │   │   └── s3_storage.go
│   │   │   └── minio/
│   │   │       └── minio_storage.go
│   │   ├── external/               # External services
│   │   │   ├── aws/
│   │   │   │   └── rekognition.go
│   │   │   ├── stripe/
│   │   │   │   └── stripe_client.go
│   │   │   └── email/
│   │   │       └── email_service.go
│   │   ├── websocket/              # WebSocket implementation
│   │   │   ├── hub.go
│   │   │   ├── client.go
│   │   │   └── message_handler.go
│   │   └── middleware/             # HTTP middleware
│   │       ├── auth_middleware.go
│   │       ├── cors_middleware.go
│   │       ├── rate_limit_middleware.go
│   │       └── logging_middleware.go
│   └── interfaces/                 # Interface adapters
│       ├── http/                    # HTTP handlers
│       │   ├── handlers/
│       │   │   ├── auth_handler.go
│       │   │   ├── user_handler.go
│       │   │   ├── photo_handler.go
│       │   │   ├── message_handler.go
│       │   │   ├── match_handler.go
│       │   │   ├── report_handler.go
│       │   │   ├── subscription_handler.go
│       │   │   └── admin_handler.go
│       │   ├── routes/
│       │   │   ├── auth_routes.go
│       │   │   ├── user_routes.go
│       │   │   ├── photo_routes.go
│   │   │   ├── message_routes.go
│       │   │   ├── match_routes.go
│       │   │   ├── report_routes.go
│   │   │   ├── subscription_routes.go
│       │   │   └── admin_routes.go
│       │   └── server.go
│       └── websocket/              # WebSocket handlers
│           └── websocket_handler.go
├── pkg/                           # Shared packages
│   ├── config/                    # Configuration management
│   │   ├── config.go
│   │   └── env.go
│   ├── logger/                    # Logging utilities
│   │   └── logger.go
│   ├── validator/                 # Validation utilities
│   │   └── validator.go
│   ├── errors/                    # Custom error types
│   │   └── errors.go
│   └── utils/                     # Utility functions
│       ├── hash.go
│       ├── jwt.go
│       └── response.go
├── migrations/                     # Database migration files
├── docs/                          # Documentation
├── scripts/                       # Build and deployment scripts
├── docker/                        # Docker configurations
│   ├── Dockerfile
│   ├── docker-compose.yml
│   └── docker-compose.dev.yml
├── .env.example                   # Environment variables template
├── go.mod                         # Go module file
├── go.sum                         # Go dependencies checksum
├── Makefile                       # Build automation
└── README.md                      # Project documentation
```

## Database Schema Design

### PostgreSQL Schema

#### Users Table
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    date_of_birth DATE NOT NULL,
    gender VARCHAR(20) NOT NULL CHECK (gender IN ('male', 'female', 'other')),
    interested_in VARCHAR(20)[] NOT NULL,
    bio TEXT,
    location_lat DECIMAL(10, 8),
    location_lng DECIMAL(11, 8),
    location_city VARCHAR(100),
    location_country VARCHAR(100),
    is_verified BOOLEAN DEFAULT FALSE,
    is_premium BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    is_banned BOOLEAN DEFAULT FALSE,
    last_active TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_location ON users USING GIST (point(location_lng, location_lat));
CREATE INDEX idx_users_active ON users(is_active, is_banned);
CREATE INDEX idx_users_last_active ON users(last_active DESC);
```

#### Photos Table
```sql
CREATE TABLE photos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    file_url VARCHAR(500) NOT NULL,
    file_key VARCHAR(255) NOT NULL,
    is_primary BOOLEAN DEFAULT FALSE,
    verification_status VARCHAR(20) DEFAULT 'pending' CHECK (verification_status IN ('pending', 'approved', 'rejected')),
    verification_reason TEXT,
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_photos_user_id ON photos(user_id);
CREATE INDEX idx_photos_primary ON photos(user_id, is_primary);
CREATE INDEX idx_photos_verification ON photos(verification_status);
```

#### User Preferences Table
```sql
CREATE TABLE user_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    age_min INTEGER DEFAULT 18,
    age_max INTEGER DEFAULT 100,
    max_distance INTEGER DEFAULT 50, -- in kilometers
    show_me BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_user_preferences_user_id ON user_preferences(user_id);
```

#### Matches Table
```sql
CREATE TABLE matches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user1_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    user2_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    matched_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_matches_users ON matches(LEAST(user1_id, user2_id), GREATEST(user1_id, user2_id));
CREATE INDEX idx_matches_user1 ON matches(user1_id);
CREATE INDEX idx_matches_user2 ON matches(user2_id);
CREATE INDEX idx_matches_active ON matches(is_active);
```

#### Swipes Table
```sql
CREATE TABLE swipes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    swiper_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    swiped_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    is_like BOOLEAN NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_swipes_unique ON swipes(swiper_id, swiped_id);
CREATE INDEX idx_swipes_swiper ON swipes(swiper_id);
CREATE INDEX idx_swipes_swiped ON swipes(swiped_id);
```

#### Conversations Table
```sql
CREATE TABLE conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    match_id UUID NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_conversations_match_id ON conversations(match_id);
```

#### Messages Table
```sql
CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    sender_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    message_type VARCHAR(20) DEFAULT 'text' CHECK (message_type IN ('text', 'image', 'gif')),
    is_read BOOLEAN DEFAULT FALSE,
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_messages_conversation ON messages(conversation_id);
CREATE INDEX idx_messages_sender ON messages(sender_id);
CREATE INDEX idx_messages_created ON messages(created_at);
CREATE INDEX idx_messages_unread ON messages(conversation_id, is_read);
```

#### Subscriptions Table
```sql
CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    stripe_subscription_id VARCHAR(255) UNIQUE,
    plan_type VARCHAR(50) NOT NULL CHECK (plan_type IN ('basic', 'premium', 'platinum')),
    status VARCHAR(50) NOT NULL CHECK (status IN ('active', 'canceled', 'past_due', 'unpaid')),
    current_period_start TIMESTAMP WITH TIME ZONE,
    current_period_end TIMESTAMP WITH TIME ZONE,
    cancel_at_period_end BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX idx_subscriptions_stripe ON subscriptions(stripe_subscription_id);
CREATE INDEX idx_subscriptions_status ON subscriptions(status);
```

#### Reports Table
```sql
CREATE TABLE reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reporter_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reported_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reason VARCHAR(100) NOT NULL CHECK (reason IN ('inappropriate_behavior', 'fake_profile', 'spam', 'harassment', 'other')),
    description TEXT,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'reviewed', 'resolved', 'dismissed')),
    reviewed_by UUID REFERENCES users(id),
    reviewed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_reports_reporter ON reports(reporter_id);
CREATE INDEX idx_reports_reported ON reports(reported_user_id);
CREATE INDEX idx_reports_status ON reports(status);
```

#### Admin Users Table
```sql
CREATE TABLE admin_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    role VARCHAR(50) DEFAULT 'admin' CHECK (role IN ('admin', 'moderator', 'super_admin')),
    is_active BOOLEAN DEFAULT TRUE,
    last_login TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_admin_users_email ON admin_users(email);
CREATE INDEX idx_admin_users_role ON admin_users(role);
```

### Redis Schema

#### Session Storage
```
Key: session:{user_id}
Type: Hash
TTL: 24 hours
Fields:
  - token: JWT token
  - refresh_token: Refresh token
  - device_info: Device information
  - last_activity: Timestamp
```

#### Online Users
```
Key: online_users
Type: Set
TTL: None
Members: user_id values
```

#### Message Cache
```
Key: messages:{conversation_id}
Type: List
TTL: 1 hour
Values: JSON serialized messages
```

#### Rate Limiting
```
Key: rate_limit:{user_id}:{endpoint}
Type: String
TTL: 1 minute/1 hour depending on endpoint
Value: Request count
```

#### Photo Verification Queue
```
Key: photo_verification_queue
Type: List
TTL: None
Values: JSON serialized photo IDs pending verification
```

## API Endpoint Specifications

### Authentication Endpoints

#### POST /api/v1/auth/register
**Description:** Register a new user account
**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "securePassword123",
  "first_name": "John",
  "last_name": "Doe",
  "date_of_birth": "1990-01-01",
  "gender": "male",
  "interested_in": ["female"]
}
```
**Response:**
```json
{
  "success": true,
  "data": {
    "user": {
      "id": "uuid",
      "email": "user@example.com",
      "first_name": "John",
      "last_name": "Doe",
      "is_verified": false,
      "created_at": "2025-01-01T00:00:00Z"
    },
    "tokens": {
      "access_token": "jwt_token",
      "refresh_token": "refresh_token",
      "expires_in": 3600
    }
  }
}
```

#### POST /api/v1/auth/login
**Description:** Authenticate user and return tokens
**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "securePassword123"
}
```
**Response:** Same as register response

#### POST /api/v1/auth/refresh
**Description:** Refresh access token using refresh token
**Request Body:**
```json
{
  "refresh_token": "refresh_token"
}
```
**Response:**
```json
{
  "success": true,
  "data": {
    "access_token": "new_jwt_token",
    "expires_in": 3600
  }
}
```

#### POST /api/v1/auth/logout
**Description:** Logout user and invalidate tokens
**Headers:** `Authorization: Bearer {token}`
**Response:**
```json
{
  "success": true,
  "message": "Successfully logged out"
}
```

### User Management Endpoints

#### GET /api/v1/users/profile
**Description:** Get current user profile
**Headers:** `Authorization: Bearer {token}`
**Response:**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "date_of_birth": "1990-01-01",
    "gender": "male",
    "interested_in": ["female"],
    "bio": "Software developer who loves hiking",
    "location": {
      "lat": 40.7128,
      "lng": -74.0060,
      "city": "New York",
      "country": "USA"
    },
    "is_verified": true,
    "is_premium": false,
    "photos": [
      {
        "id": "uuid",
        "url": "https://cdn.example.com/photo1.jpg",
        "is_primary": true,
        "verification_status": "approved"
      }
    ],
    "preferences": {
      "age_min": 25,
      "age_max": 35,
      "max_distance": 50,
      "show_me": true
    }
  }
}
```

#### PUT /api/v1/users/profile
**Description:** Update user profile
**Headers:** `Authorization: Bearer {token}`
**Request Body:**
```json
{
  "first_name": "John",
  "last_name": "Smith",
  "bio": "Updated bio",
  "location": {
    "lat": 40.7128,
    "lng": -74.0060,
    "city": "New York",
    "country": "USA"
  },
  "preferences": {
    "age_min": 25,
    "age_max": 35,
    "max_distance": 50,
    "show_me": true
  }
}
```

#### DELETE /api/v1/users/account
**Description:** Delete user account
**Headers:** `Authorization: Bearer {token}`
**Response:**
```json
{
  "success": true,
  "message": "Account successfully deleted"
}
```

### Photo Management Endpoints

#### POST /api/v1/photos
**Description:** Upload a new photo
**Headers:** `Authorization: Bearer {token}`
**Request:** `multipart/form-data` with file
**Response:**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "url": "https://cdn.example.com/photo.jpg",
    "is_primary": false,
    "verification_status": "pending"
  }
}
```

#### PUT /api/v1/photos/{photo_id}/primary
**Description:** Set photo as primary
**Headers:** `Authorization: Bearer {token}`
**Response:**
```json
{
  "success": true,
  "message": "Photo set as primary"
}
```

#### DELETE /api/v1/photos/{photo_id}
**Description:** Delete a photo
**Headers:** `Authorization: Bearer {token}`
**Response:**
```json
{
  "success": true,
  "message": "Photo deleted successfully"
}
```

### Matching Endpoints

#### GET /api/v1/matches/potential
**Description:** Get potential matches for swiping
**Headers:** `Authorization: Bearer {token}`
**Query Parameters:**
- `limit`: Number of profiles to return (default: 10)
- `offset`: Pagination offset (default: 0)
**Response:**
```json
{
  "success": true,
  "data": {
    "profiles": [
      {
        "id": "uuid",
        "first_name": "Jane",
        "age": 28,
        "bio": "Love traveling and photography",
        "location": {
          "city": "New York",
          "distance": 5.2
        },
        "photos": [
          {
            "url": "https://cdn.example.com/photo1.jpg",
            "is_primary": true
          }
        ]
      }
    ],
    "pagination": {
      "total": 100,
      "limit": 10,
      "offset": 0
    }
  }
}
```

#### POST /api/v1/matches/swipe
**Description:** Swipe on a user profile
**Headers:** `Authorization: Bearer {token}`
**Request Body:**
```json
{
  "swiped_user_id": "uuid",
  "is_like": true
}
```
**Response:**
```json
{
  "success": true,
  "data": {
    "is_match": true,
    "match_id": "uuid"
  }
}
```

#### GET /api/v1/matches
**Description:** Get user's matches
**Headers:** `Authorization: Bearer {token}`
**Query Parameters:**
- `limit`: Number of matches to return (default: 20)
- `offset`: Pagination offset (default: 0)
**Response:**
```json
{
  "success": true,
  "data": {
    "matches": [
      {
        "id": "uuid",
        "user": {
          "id": "uuid",
          "first_name": "Jane",
          "age": 28,
          "photos": [
            {
              "url": "https://cdn.example.com/photo1.jpg",
              "is_primary": true
            }
          ]
        },
        "matched_at": "2025-01-01T00:00:00Z",
        "conversation_id": "uuid"
      }
    ],
    "pagination": {
      "total": 50,
      "limit": 20,
      "offset": 0
    }
  }
}
```

### Messaging Endpoints

#### GET /api/v1/messages/conversations
**Description:** Get user's conversations
**Headers:** `Authorization: Bearer {token}`
**Query Parameters:**
- `limit`: Number of conversations to return (default: 20)
- `offset`: Pagination offset (default: 0)
**Response:**
```json
{
  "success": true,
  "data": {
    "conversations": [
      {
        "id": "uuid",
        "match": {
          "user": {
            "id": "uuid",
            "first_name": "Jane",
            "photos": [
              {
                "url": "https://cdn.example.com/photo1.jpg",
                "is_primary": true
              }
            ]
          }
        },
        "last_message": {
          "content": "Hey! How are you?",
          "sender_id": "uuid",
          "created_at": "2025-01-01T00:00:00Z"
        },
        "unread_count": 2,
        "updated_at": "2025-01-01T00:00:00Z"
      }
    ],
    "pagination": {
      "total": 10,
      "limit": 20,
      "offset": 0
    }
  }
}
```

#### GET /api/v1/messages/conversations/{conversation_id}
**Description:** Get messages in a conversation
**Headers:** `Authorization: Bearer {token}`
**Query Parameters:**
- `limit`: Number of messages to return (default: 50)
- `offset`: Pagination offset (default: 0)
**Response:**
```json
{
  "success": true,
  "data": {
    "messages": [
      {
        "id": "uuid",
        "sender_id": "uuid",
        "content": "Hey! How are you?",
        "message_type": "text",
        "is_read": true,
        "created_at": "2025-01-01T00:00:00Z"
      }
    ],
    "pagination": {
      "total": 100,
      "limit": 50,
      "offset": 0
    }
  }
}
```

#### POST /api/v1/messages/conversations/{conversation_id}
**Description:** Send a message
**Headers:** `Authorization: Bearer {token}`
**Request Body:**
```json
{
  "content": "Hey! How are you?",
  "message_type": "text"
}
```
**Response:**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "sender_id": "uuid",
    "content": "Hey! How are you?",
    "message_type": "text",
    "is_read": false,
    "created_at": "2025-01-01T00:00:00Z"
  }
}
```

#### PUT /api/v1/messages/{message_id}/read
**Description:** Mark message as read
**Headers:** `Authorization: Bearer {token}`
**Response:**
```json
{
  "success": true,
  "message": "Message marked as read"
}
```

### Subscription Endpoints

#### GET /api/v1/subscriptions/plans
**Description:** Get available subscription plans
**Response:**
```json
{
  "success": true,
  "data": {
    "plans": [
      {
        "id": "premium",
        "name": "Premium",
        "price": 9.99,
        "currency": "USD",
        "interval": "month",
        "features": [
          "Unlimited swipes",
          "See who likes you",
          "5 super likes per day",
          "1 boost per month"
        ]
      },
      {
        "id": "platinum",
        "name": "Platinum",
        "price": 19.99,
        "currency": "USD",
        "interval": "month",
        "features": [
          "All Premium features",
          "Unlimited super likes",
          "Weekly boosts",
          "Priority support"
        ]
      }
    ]
  }
}
```

#### POST /api/v1/subscriptions
**Description:** Create a new subscription
**Headers:** `Authorization: Bearer {token}`
**Request Body:**
```json
{
  "plan_id": "premium",
  "payment_method_id": "pm_stripe_token"
}
```
**Response:**
```json
{
  "success": true,
  "data": {
    "subscription_id": "uuid",
    "client_secret": "stripe_client_secret",
    "status": "pending"
  }
}
```

#### GET /api/v1/subscriptions/current
**Description:** Get current user subscription
**Headers:** `Authorization: Bearer {token}`
**Response:**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "plan_type": "premium",
    "status": "active",
    "current_period_start": "2025-01-01T00:00:00Z",
    "current_period_end": "2025-02-01T00:00:00Z",
    "cancel_at_period_end": false
  }
}
```

#### POST /api/v1/subscriptions/cancel
**Description:** Cancel subscription
**Headers:** `Authorization: Bearer {token}`
**Response:**
```json
{
  "success": true,
  "message": "Subscription will be canceled at the end of the billing period"
}
```

#### POST /api/v1/subscriptions/webhook
**Description:** Handle Stripe webhooks
**Request Headers:** `Stripe-Signature`
**Response:** HTTP 200 OK

### Reporting Endpoints

#### POST /api/v1/reports
**Description:** Report a user
**Headers:** `Authorization: Bearer {token}`
**Request Body:**
```json
{
  "reported_user_id": "uuid",
  "reason": "inappropriate_behavior",
  "description": "User sent inappropriate messages"
}
```
**Response:**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "status": "pending"
  }
}
```

### Admin Endpoints

#### POST /api/v1/admin/auth/login
**Description:** Admin login
**Request Body:**
```json
{
  "email": "admin@example.com",
  "password": "adminPassword123"
}
```

#### GET /api/v1/admin/reports
**Description:** Get all reports (admin only)
**Headers:** `Authorization: Bearer {admin_token}`
**Query Parameters:**
- `status`: Filter by status (pending, reviewed, resolved, dismissed)
- `limit`: Number of reports to return (default: 20)
- `offset`: Pagination offset (default: 0)

#### PUT /api/v1/admin/reports/{report_id}/review
**Description:** Review a report (admin only)
**Headers:** `Authorization: Bearer {admin_token}`
**Request Body:**
```json
{
  "status": "resolved",
  "action": "ban_user"
}
```

#### GET /api/v1/admin/users
**Description:** Get all users (admin only)
**Headers:** `Authorization: Bearer {admin_token}`
**Query Parameters:**
- `status`: Filter by status (active, banned, verified)
- `limit`: Number of users to return (default: 20)
- `offset`: Pagination offset (default: 0)

#### PUT /api/v1/admin/users/{user_id}/ban
**Description:** Ban a user (admin only)
**Headers:** `Authorization: Bearer {admin_token}`
**Request Body:**
```json
{
  "reason": "Violation of community guidelines"
}
```

### WebSocket Endpoints

#### WS /api/v1/ws
**Description:** WebSocket connection for real-time messaging
**Headers:** `Authorization: Bearer {token}`
**Message Format:**
```json
{
  "type": "message",
  "data": {
    "conversation_id": "uuid",
    "content": "Hello!",
    "message_type": "text"
  }
}
```

**Response Format:**
```json
{
  "type": "message",
  "data": {
    "id": "uuid",
    "conversation_id": "uuid",
    "sender_id": "uuid",
    "content": "Hello!",
    "message_type": "text",
    "created_at": "2025-01-01T00:00:00Z"
  }
}
```

## Security Architecture

### JWT Authentication Flow

1. **User Registration/Login:**
   - User provides credentials
   - Server validates credentials
   - Server generates JWT access token (15 minutes expiry)
   - Server generates refresh token (7 days expiry)
   - Both tokens are returned to client

2. **Token Refresh:**
   - Client sends refresh token to `/api/v1/auth/refresh`
   - Server validates refresh token
   - Server generates new access token
   - New access token is returned to client

3. **Protected Routes:**
   - Client includes access token in `Authorization: Bearer {token}` header
   - Middleware validates token signature and expiry
   - Middleware extracts user claims and sets context

4. **Token Storage:**
   - Access tokens stored in memory (short-lived)
   - Refresh tokens stored in secure HTTP-only cookies
   - Redis session storage for additional validation

### Rate Limiting Strategy

#### Endpoint-Specific Limits:
- **Authentication endpoints:** 5 requests per minute per IP
- **Photo upload:** 10 requests per hour per user
- **Messaging:** 60 messages per minute per user
- **Matching/Swiping:** 100 swipes per hour per user
- **General API:** 1000 requests per hour per user

#### Implementation:
- Redis-based rate limiting using sliding window algorithm
- Rate limit headers included in responses:
  - `X-RateLimit-Limit`: Request limit
  - `X-RateLimit-Remaining`: Remaining requests
  - `X-RateLimit-Reset`: Reset time (Unix timestamp)

### Data Protection Measures

#### Password Security:
- bcrypt hashing with cost factor 12
- Minimum password requirements: 8 characters, uppercase, lowercase, number, special character
- Password strength validation during registration

#### PII Protection:
- Email addresses encrypted at rest
- Location data precision reduction (3 decimal places)
- Automatic data anonymization for deleted accounts after 30 days

#### API Security:
- HTTPS enforcement (HSTS)
- CORS configuration for frontend domains
- Input validation and sanitization
- SQL injection prevention through parameterized queries
- XSS prevention through output encoding

#### File Upload Security:
- File type validation (images only)
- File size limits (5MB per photo)
- Virus scanning integration
- Automatic image optimization and compression

### Privacy Controls

#### User Consent:
- GDPR compliance with explicit consent mechanisms
- Data processing transparency
- Right to data portability
- Right to be forgotten

#### Data Minimization:
- Collect only necessary user data
- Automatic cleanup of unused data
- Limited data retention periods

## External Service Integrations

### AWS Rekognition Integration

#### Photo Verification Workflow:
1. User uploads photo via API
2. Photo is stored in S3/MinIO
3. Photo ID is added to Redis queue for verification
4. Background worker processes queue:
   - Calls AWS Rekognition `DetectModerationLabels`
   - Calls AWS Rekognition `CompareFaces` for profile verification
   - Analyzes image quality and appropriateness
5. Results are stored in database
6. User is notified of verification status

#### Implementation Details:
```go
type RekognitionService struct {
    client *rekognition.Rekognition
    region string
}

func (r *RekognitionService) AnalyzePhoto(imageKey string) (*PhotoAnalysis, error) {
    // Detect inappropriate content
    moderationInput := &rekognition.DetectModerationLabelsInput{
        Image: &rekognition.Image{
            S3Object: &rekognition.S3Object{
                Bucket: aws.String(r.bucket),
                Name:   aws.String(imageKey),
            },
        },
        MinConfidence: aws.Float64(75.0),
    }
    
    moderationResult, err := r.client.DetectModerationLabels(moderationInput)
    if err != nil {
        return nil, err
    }
    
    // Analyze results and return structured data
    return &PhotoAnalysis{
        IsAppropriate: len(moderationResult.ModerationLabels) == 0,
        Confidence:    calculateConfidence(moderationResult.ModerationLabels),
        Labels:       extractLabels(moderationResult.ModerationLabels),
    }, nil
}
```

#### Error Handling:
- Automatic retry with exponential backoff
- Fallback to manual review if AWS service unavailable
- Cost monitoring and limits

### Stripe Integration

#### Subscription Management:
1. User selects subscription plan
2. Frontend creates Stripe payment method
3. Backend creates Stripe customer and subscription
4. Webhook handlers update local database
5. Real-time subscription status synchronization

#### Implementation Details:
```go
type StripeService struct {
    client *stripe.Client
    webhookSecret string
}

func (s *StripeService) CreateSubscription(userID string, planID string, paymentMethodID string) (*stripe.Subscription, error) {
    // Create or retrieve customer
    customer, err := s.client.Customers.New(&stripe.CustomerParams{
        PaymentMethod: stripe.String(paymentMethodID),
        Email:         stripe.String(getUserEmail(userID)),
        Metadata: map[string]string{
            "user_id": userID,
        },
    })
    
    if err != nil {
        return nil, err
    }
    
    // Create subscription
    subscription, err := s.client.Subscriptions.New(&stripe.SubscriptionParams{
        Customer: stripe.String(customer.ID),
        Items: []*stripe.SubscriptionItemsParams{
            {
                Price: stripe.String(getPlanPriceID(planID)),
            },
        },
    })
    
    return subscription, err
}
```

#### Webhook Events:
- `customer.subscription.created`
- `customer.subscription.updated`
- `customer.subscription.deleted`
- `invoice.payment_succeeded`
- `invoice.payment_failed`

### S3/MinIO Integration

#### File Storage Strategy:
1. User uploads photo via multipart form
2. Backend validates file type and size
3. File is uploaded to S3/MinIO with unique key
4. Database record is created with file metadata
5. CDN URL is returned to client

#### Implementation Details:
```go
type StorageService interface {
    UploadFile(ctx context.Context, file io.Reader, key string, contentType string) (string, error)
    DeleteFile(ctx context.Context, key string) error
    GetFileURL(ctx context.Context, key string) (string, error)
}

type S3Storage struct {
    client *s3.Client
    bucket string
    region string
}

func (s *S3Storage) UploadFile(ctx context.Context, file io.Reader, key string, contentType string) (string, error) {
    _, err := s.client.PutObject(ctx, &s3.PutObjectInput{
        Bucket:      aws.String(s.bucket),
        Key:         aws.String(key),
        Body:        file,
        ContentType: aws.String(contentType),
        ACL:         types.ObjectCannedACLPrivate,
    })
    
    if err != nil {
        return "", err
    }
    
    return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucket, s.region, key), nil
}
```

#### File Organization:
```
s3://dating-app-bucket/
├── photos/
│   ├── original/
│   │   ├── user1/
│   │   │   ├── photo1.jpg
│   │   │   └── photo2.jpg
│   │   └── user2/
│   │       └── photo1.jpg
│   └── thumbnails/
│       ├── user1/
│       │   ├── photo1_thumb.jpg
│       │   └── photo2_thumb.jpg
│       └── user2/
│           └── photo1_thumb.jpg
└── temp/
    └── uploads/
        └── temp_file_123456.jpg
```

## Technology Stack

### Core Technologies

#### Backend Framework
- **Gin**: HTTP web framework for high-performance routing
- **GORM**: ORM for database operations with PostgreSQL
- **sqlx**: Additional database operations for complex queries

#### Database
- **PostgreSQL 14+**: Primary database with JSONB support
- **Redis 6+**: Caching, sessions, and real-time features
- **Flyway**: Database migration management

#### Real-time Communication
- **Gorilla WebSocket**: WebSocket implementation
- **Go-Redis**: Redis client for pub/sub and caching

### Go Libraries and Frameworks

#### Core Libraries
```go
// Web Framework
github.com/gin-gonic/gin v1.9.1

// Database
gorm.io/gorm v1.25.4
gorm.io/driver/postgres v1.5.2
github.com/jmoiron/sqlx v1.3.5

// Redis
github.com/go-redis/redis/v8 v8.11.5

// Validation
github.com/go-playground/validator/v10 v10.15.1

// JWT
github.com/golang-jwt/jwt/v5 v5.0.0

// Configuration
github.com/spf13/viper v1.16.0

// Logging
github.com/sirupsen/logrus v1.9.3
go.uber.org/zap v1.25.0

// HTTP Client
github.com/go-resty/resty/v2 v2.7.0

// UUID
github.com/google/uuid v1.3.0

// Password Hashing
golang.org/x/crypto v0.12.0
```

#### External Service Libraries
```go
// AWS SDK
github.com/aws/aws-sdk-go-v2 v1.21.0
github.com/aws/aws-sdk-go-v2/service/rekognition v1.21.0
github.com/aws/aws-sdk-go-v2/service/s3 v1.38.5

// Stripe
github.com/stripe/stripe-go/v76 v76.15.0

// MinIO
github.com/minio/minio-go/v7 v7.0.63

// Email
github.com/sendgrid/sendgrid-go v3.10.0
```

#### Development Tools
```go
// Testing
github.com/stretchr/testify v1.8.4
github.com/golang/mock v1.6.0

// Code Generation
github.com/swaggo/gin-swagger v1.6.0
github.com/swaggo/swag v1.16.1

// Profiling
github.com/pkg/profile v1.6.0

// Metrics
github.com/prometheus/client_golang v1.16.0
```

### Development Tools

#### Code Quality
- **golangci-lint**: Go linting and static analysis
- **gofmt**: Code formatting
- **goimports**: Import management
- **gosec**: Security vulnerability scanning

#### Testing
- **testify**: Assertion library
- **gomock**: Mocking framework
- **dockertest**: Database testing with Docker

#### Documentation
- **swaggo**: OpenAPI/Swagger documentation generation
- **godoc**: Go documentation

### Infrastructure Tools

#### Containerization
- **Docker**: Application containerization
- **Docker Compose**: Local development environment

#### CI/CD
- **GitHub Actions**: Continuous integration and deployment
- **Make**: Build automation

#### Monitoring
- **Prometheus**: Metrics collection
- **Grafana**: Metrics visualization
- **Jaeger**: Distributed tracing

## Deployment Architecture

### Local Development Setup

#### Docker Compose Configuration
```yaml
version: '3.8'

services:
  app:
    build:
      context: .
      dockerfile: docker/Dockerfile.dev
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=dating_user
      - DB_PASSWORD=dating_pass
      - DB_NAME=dating_db
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - AWS_REGION=us-east-1
      - AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID}
      - AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY}
      - STRIPE_SECRET_KEY=${STRIPE_SECRET_KEY}
      - STRIPE_WEBHOOK_SECRET=${STRIPE_WEBHOOK_SECRET}
    volumes:
      - .:/app
    depends_on:
      - postgres
      - redis
    networks:
      - dating-network

  postgres:
    image: postgres:14-alpine
    environment:
      - POSTGRES_USER=dating_user
      - POSTGRES_PASSWORD=dating_pass
      - POSTGRES_DB=dating_db
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./migrations:/docker-entrypoint-initdb.d
    networks:
      - dating-network

  redis:
    image: redis:6-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    networks:
      - dating-network

  minio:
    image: minio/minio:latest
    ports:
      - "9000:9000"
      - "9001:9001"
    environment:
      - MINIO_ROOT_USER=minioadmin
      - MINIO_ROOT_PASSWORD=minioadmin
    command: server /data --console-address ":9001"
    volumes:
      - minio_data:/data
    networks:
      - dating-network

volumes:
  postgres_data:
  redis_data:
  minio_data:

networks:
  dating-network:
    driver: bridge
```

#### Development Environment Variables
```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=dating_user
DB_PASSWORD=dating_pass
DB_NAME=dating_db
DB_SSL_MODE=disable

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT
JWT_SECRET=your-super-secret-jwt-key
JWT_ACCESS_TOKEN_EXPIRY=15m
JWT_REFRESH_TOKEN_EXPIRY=168h

# AWS
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your-aws-access-key
AWS_SECRET_ACCESS_KEY=your-aws-secret-key
S3_BUCKET=dating-app-dev

# Stripe
STRIPE_SECRET_KEY=sk_test_your-stripe-secret-key
STRIPE_WEBHOOK_SECRET=whsec_your-webhook-secret

# Email
SENDGRID_API_KEY=your-sendgrid-api-key

# Application
APP_ENV=development
APP_PORT=8080
APP_HOST=localhost
LOG_LEVEL=debug

# Rate Limiting
RATE_LIMIT_REQUESTS_PER_MINUTE=1000
RATE_LIMIT_REQUESTS_PER_HOUR=10000
```

### Application Configuration

#### Configuration Structure
```go
type Config struct {
    App      AppConfig      `mapstructure:"app"`
    Database DatabaseConfig `mapstructure:"database"`
    Redis    RedisConfig    `mapstructure:"redis"`
    JWT      JWTConfig      `mapstructure:"jwt"`
    AWS      AWSConfig      `mapstructure:"aws"`
    Stripe   StripeConfig   `mapstructure:"stripe"`
    Email    EmailConfig    `mapstructure:"email"`
    RateLimit RateLimitConfig `mapstructure:"rate_limit"`
}

type AppConfig struct {
    Env  string `mapstructure:"env"`
    Port int    `mapstructure:"port"`
    Host string `mapstructure:"host"`
}

type DatabaseConfig struct {
    Host     string `mapstructure:"host"`
    Port     int    `mapstructure:"port"`
    User     string `mapstructure:"user"`
    Password string `mapstructure:"password"`
    DBName   string `mapstructure:"db_name"`
    SSLMode  string `mapstructure:"ssl_mode"`
}

type RedisConfig struct {
    Host     string `mapstructure:"host"`
    Port     int    `mapstructure:"port"`
    Password string `mapstructure:"password"`
    DB       int    `mapstructure:"db"`
}

type JWTConfig struct {
    Secret              string        `mapstructure:"secret"`
    AccessTokenExpiry   time.Duration `mapstructure:"access_token_expiry"`
    RefreshTokenExpiry  time.Duration `mapstructure:"refresh_token_expiry"`
}
```

### Database Migration Strategy

#### Migration File Structure
```
migrations/
├── 001_create_users_table.up.sql
├── 001_create_users_table.down.sql
├── 002_create_photos_table.up.sql
├── 002_create_photos_table.down.sql
├── 003_create_user_preferences_table.up.sql
├── 003_create_user_preferences_table.down.sql
└── ...
```

#### Migration Tool Integration
```go
// Using golang-migrate for database migrations
func runMigrations(config *Config) error {
    m, err := migrate.New(
        fmt.Sprintf("file://%s", "migrations"),
        fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
            config.Database.User,
            config.Database.Password,
            config.Database.Host,
            config.Database.Port,
            config.Database.DBName,
            config.Database.SSLMode,
        ),
    )
    
    if err != nil {
        return err
    }
    
    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        return err
    }
    
    return nil
}
```

### Development Workflow

#### Makefile Commands
```makefile
.PHONY: help build run test clean docker-up docker-down migrate-up migrate-down

help:
	@echo "Available commands:"
	@echo "  build        - Build the application"
	@echo "  run          - Run the application"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"
	@echo "  docker-up    - Start development environment"
	@echo "  docker-down  - Stop development environment"
	@echo "  migrate-up   - Run database migrations"
	@echo "  migrate-down - Rollback database migrations"

build:
	go build -o bin/api cmd/api/main.go

run:
	go run cmd/api/main.go

test:
	go test -v ./...

clean:
	rm -rf bin/

docker-up:
	docker-compose -f docker/docker-compose.dev.yml up -d

docker-down:
	docker-compose -f docker/docker-compose.dev.yml down

migrate-up:
	migrate -path migrations -database "postgres://dating_user:dating_pass@localhost:5432/dating_db?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgres://dating_user:dating_pass@localhost:5432/dating_db?sslmode=disable" down
```

#### Local Development Steps
1. **Setup Environment:**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

2. **Start Services:**
   ```bash
   make docker-up
   ```

3. **Run Migrations:**
   ```bash
   make migrate-up
   ```

4. **Start Application:**
   ```bash
   make run
   ```

5. **Run Tests:**
   ```bash
   make test
   ```

### Performance Considerations

#### Database Optimization
- Proper indexing on frequently queried columns
- Connection pooling with pgxpool
- Query optimization and N+1 prevention
- Read replicas for scaling read operations

#### Caching Strategy
- Redis for session storage
- Application-level caching for frequently accessed data
- CDN integration for static assets
- Database query result caching

#### API Performance
- Request/response compression
- Pagination for large datasets
- Efficient JSON serialization
- Background job processing for heavy operations

This architecture provides a comprehensive foundation for building a scalable, secure, and maintainable dating application backend using Go and modern best practices.
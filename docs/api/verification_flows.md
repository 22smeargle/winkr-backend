# Verification API Flows

This document describes the complete verification flows for the Winkr dating application, including AI-powered selfie verification, document verification, and admin review processes.

## Overview

The verification system provides two levels of user verification:
1. **Selfie Verification** (Level 1) - Basic photo verification with AI analysis
2. **Document Verification** (Level 2) - Advanced verification with government-issued documents

Both verification types use AWS Rekognition for AI analysis and include fraud detection, rate limiting, and manual review capabilities.

## Selfie Verification Flow

### Step 1: Request Selfie Verification

**Endpoint:** `POST /api/v1/verify/selfie/request`

**Purpose:** Initiates the selfie verification process and returns a secure upload URL.

**Request:**
```json
{}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "verification_id": "550e8400-e29b-41d4-a716-446655440000",
    "upload_url": "https://storage.example.com/upload/abc123",
    "expires_at": "2025-12-01T12:00:00Z"
  }
}
```

**Security Features:**
- Requires valid JWT token
- Rate limited (3 attempts per day, 10 per month)
- IP tracking for fraud detection
- Upload URL expires in 15 minutes

### Step 2: Upload Selfie Photo

**Action:** Upload photo to the provided secure URL

**Requirements:**
- File format: JPEG, PNG, or WebP
- Maximum file size: 5MB
- Photo must be taken within last 24 hours (EXIF data validation)
- Must contain a clear, front-facing face

### Step 3: Submit Selfie Verification

**Endpoint:** `POST /api/v1/verify/selfie/submit`

**Purpose:** Submits the uploaded selfie for AI-powered verification.

**Request:** `multipart/form-data`
- `photo`: Binary image file
- `verification_id`: UUID from step 1

**AI Processing:**
1. **Face Detection** - Detects faces in the image
2. **Face Analysis** - Analyzes facial features, emotions, age, gender
3. **NSFW Detection** - Checks for inappropriate content
4. **Liveness Detection** - Anti-spoofing checks (blink, smile, movement)
5. **Quality Assessment** - Evaluates image quality and clarity

**Response:**
```json
{
  "success": true,
  "data": {
    "verification_id": "550e8400-e29b-41d4-a716-446655440000",
    "status": "pending",
    "ai_score": 0.92,
    "ai_details": {
      "face_detected": true,
      "confidence": 0.95,
      "nsfw_detected": false,
      "liveness_detected": true
    },
    "created_at": "2025-12-01T10:00:00Z"
  }
}
```

**AI Confidence Thresholds:**
- Face detection confidence: > 0.80
- Liveness confidence: > 0.90
- NSFW detection: < 0.70
- Overall similarity: > 0.85

### Step 4: Check Verification Status

**Endpoint:** `GET /api/v1/verify/selfie/status`

**Purpose:** Retrieves the current status of selfie verification.

**Response:**
```json
{
  "success": true,
  "data": {
    "verification_id": "550e8400-e29b-41d4-a716-446655440000",
    "status": "approved",
    "ai_score": 0.92,
    "ai_details": {
      "face_detected": true,
      "confidence": 0.95,
      "nsfw_detected": false,
      "liveness_detected": true
    },
    "reviewed_at": "2025-12-01T12:30:00Z",
    "expires_at": "2024-12-01T10:00:00Z",
    "created_at": "2025-12-01T10:00:00Z",
    "updated_at": "2025-12-01T12:30:00Z"
  }
}
```

**Status Values:**
- `pending` - Under AI processing or manual review
- `approved` - Verification passed, badge awarded
- `rejected` - Verification failed, reason provided

## Document Verification Flow

### Step 1: Request Document Verification

**Endpoint:** `POST /api/v1/verify/document/request`

**Purpose:** Initiates document verification and returns secure upload URL.

**Request:**
```json
{
  "document_type": "passport"
}
```

**Supported Document Types:**
- `id_card` - Government-issued ID card
- `passport` - Passport document
- `driver_license` - Driver's license

**Response:**
```json
{
  "success": true,
  "data": {
    "verification_id": "550e8400-e29b-41d4-a716-446655440001",
    "upload_url": "https://storage.example.com/upload/def456",
    "expires_at": "2025-12-01T12:00:00Z"
  }
}
```

### Step 2: Upload Document Photo

**Action:** Upload document photo to the provided secure URL

**Requirements:**
- File format: JPEG, PNG, or WebP
- Maximum file size: 10MB
- Document must be clearly visible and readable
- All four corners of document should be visible

### Step 3: Submit Document Verification

**Endpoint:** `POST /api/v1/verify/document/submit`

**Purpose:** Submits the uploaded document for AI-powered verification and OCR processing.

**Request:** `multipart/form-data`
- `photo`: Binary image file
- `verification_id`: UUID from step 1

**AI Processing:**
1. **Document Type Detection** - Identifies document type (ID card, passport, etc.)
2. **OCR Processing** - Extracts text and data from document
3. **Field Validation** - Validates extracted data format and consistency
4. **Authenticity Checks** - Detects signs of tampering or forgery
5. **Quality Assessment** - Evaluates image clarity and readability

**Response:**
```json
{
  "success": true,
  "data": {
    "verification_id": "550e8400-e29b-41d4-a716-446655440001",
    "status": "pending",
    "document_type": "passport",
    "extracted_data": {
      "document_number": "P123456789",
      "name": "John Doe",
      "nationality": "US",
      "date_of_birth": "1990-01-01",
      "expiry_date": "2030-01-01"
    },
    "ai_score": 0.87,
    "ai_details": {
      "document_detected": true,
      "confidence": 0.87,
      "ocr_confidence": 0.92,
      "validation_passed": true
    },
    "created_at": "2025-12-01T10:00:00Z"
  }
}
```

**OCR Extraction Fields:**
- **ID Card:** Document number, name, date of birth, expiry date, address
- **Passport:** Passport number, name, nationality, date of birth, expiry date
- **Driver's License:** License number, name, date of birth, expiry date, address

### Step 4: Check Document Verification Status

**Endpoint:** `GET /api/v1/verify/document/status`

**Purpose:** Retrieves the current status of document verification.

**Response:**
```json
{
  "success": true,
  "data": {
    "verification_id": "550e8400-e29b-41d4-a716-446655440001",
    "status": "approved",
    "document_type": "passport",
    "extracted_data": {
      "document_number": "P123456789",
      "name": "John Doe",
      "nationality": "US",
      "date_of_birth": "1990-01-01",
      "expiry_date": "2030-01-01"
    },
    "ai_score": 0.87,
    "ai_details": {
      "document_detected": true,
      "confidence": 0.87,
      "ocr_confidence": 0.92,
      "validation_passed": true
    },
    "reviewed_at": "2025-12-01T12:30:00Z",
    "expires_at": "2026-12-01T10:00:00Z",
    "created_at": "2025-12-01T10:00:00Z",
    "updated_at": "2025-12-01T12:30:00Z"
  }
}
```

## Admin Review Flow

### Step 1: Get Pending Verifications

**Endpoint:** `GET /api/v1/admin/verifications`

**Purpose:** Retrieves a list of verifications pending admin review.

**Query Parameters:**
- `limit` (optional): Maximum number of verifications (default: 20, max: 100)
- `offset` (optional): Number of verifications to skip (default: 0)
- `status` (optional): Filter by status (pending, approved, rejected)
- `type` (optional): Filter by type (selfie, document)

**Response:**
```json
{
  "success": true,
  "data": {
    "verifications": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "user_id": "550e8400-e29b-41d4-a716-446655440001",
        "type": "selfie",
        "status": "pending",
        "ai_score": 0.92,
        "created_at": "2025-12-01T10:00:00Z",
        "user": {
          "id": "550e8400-e29b-41d4-a716-446655440001",
          "first_name": "John",
          "last_name": "Doe",
          "email": "john.doe@example.com"
        }
      }
    ],
    "total": 15,
    "limit": 20,
    "offset": 0
  }
}
```

### Step 2: Get Verification Details

**Endpoint:** `GET /api/v1/admin/verifications/{id}`

**Purpose:** Retrieves detailed information about a specific verification.

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "user_id": "550e8400-e29b-41d4-a716-446655440001",
    "type": "selfie",
    "status": "pending",
    "photo_url": "https://storage.example.com/photos/verification.jpg",
    "ai_score": 0.92,
    "ai_details": {
      "face_detected": true,
      "confidence": 0.95,
      "nsfw_detected": false,
      "liveness_detected": true
    },
    "created_at": "2025-12-01T10:00:00Z",
    "updated_at": "2025-12-01T10:00:00Z",
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "first_name": "John",
      "last_name": "Doe",
      "email": "john.doe@example.com",
      "verification_level": 0
    },
    "reviewer": null
  }
}
```

### Step 3: Approve Verification

**Endpoint:** `POST /api/v1/admin/verifications/{id}/approve`

**Purpose:** Approves a verification and awards verification badge.

**Request:**
```json
{
  "reason": "User verified successfully"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "verification_id": "550e8400-e29b-41d4-a716-446655440000",
    "status": "approved",
    "reviewed_at": "2025-12-01T12:30:00Z",
    "message": "Verification approved successfully"
  }
}
```

**Approval Actions:**
- Updates verification status to "approved"
- Awards verification badge to user
- Updates user verification level
- Sets verification expiration date
- Records admin reviewer and timestamp

### Step 4: Reject Verification

**Endpoint:** `POST /api/v1/admin/verifications/{id}/reject`

**Purpose:** Rejects a verification with a specific reason.

**Request:**
```json
{
  "reason": "Photo quality is too low"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "verification_id": "550e8400-e29b-41d4-a716-446655440000",
    "status": "rejected",
    "reviewed_at": "2025-12-01T12:30:00Z",
    "message": "Verification rejected successfully"
  }
}
```

**Rejection Actions:**
- Updates verification status to "rejected"
- Records rejection reason for user feedback
- Does not affect user verification level
- Records admin reviewer and timestamp

## AI Processing Details

### Face Analysis

**AWS Rekognition Features:**
- Face detection and bounding boxes
- Facial landmark detection (eyes, nose, mouth)
- Emotion analysis (happy, sad, angry, surprised, etc.)
- Age range estimation
- Gender detection
- Pose and orientation analysis
- Image quality assessment

**Confidence Thresholds:**
- Face detection: > 0.80
- Feature analysis: > 0.75
- Quality assessment: > 0.70

### Face Comparison

**Process:**
1. Extract face features from verification photo
2. Compare with reference photo (if available)
3. Calculate similarity score (0.0 - 1.0)
4. Determine match based on threshold

**Similarity Threshold:** 0.85 (configurable)

### Liveness Detection

**Anti-Spoofing Techniques:**
- Eye blink detection
- Smile detection
- Head movement analysis
- Texture analysis for photo vs. live person
- Challenge-response verification

**Confidence Threshold:** 0.90 (configurable)

### Content Moderation

**NSFW Detection:**
- Explicit content detection
- Suggestive content analysis
- Inappropriate image detection

**NSFW Threshold:** 0.70 (configurable)

### Document Processing

**OCR Capabilities:**
- Text extraction from document images
- Field recognition and categorization
- Data validation and formatting
- Multi-language support

**Document Validation:**
- Format validation for each document type
- Consistency checks between fields
- Expiry date validation
- Authenticity pattern matching

## Security Features

### Rate Limiting

**User-Level Limits:**
- Selfie verification: 3 attempts per day, 10 per month
- Document verification: 3 attempts per day, 10 per month
- Cooldown period: 24 hours between attempts

**IP-Level Limits:**
- Maximum 10 verification attempts per hour per IP
- Temporary blocks for suspicious patterns

### Fraud Detection

**IP Tracking:**
- Monitor verification attempts per IP address
- Detect VPN and proxy usage patterns
- Block suspicious IP ranges

**Device Tracking:**
- Track user agent and device fingerprint
- Detect multiple accounts from same device
- Flag unusual device patterns

**Behavioral Analysis:**
- Monitor verification attempt patterns
- Detect automated submission attempts
- Flag rapid-fire verification requests

### Data Protection

**Encryption:**
- All verification photos encrypted at rest
- Secure upload URLs with short expiration
- Signed URLs for access control

**Privacy:**
- Automatic deletion of rejected photos after 30 days
- Limited access to verification data
- GDPR compliance for user data

## Verification Levels

### Level 0: None
- No verification completed
- Basic platform access only

### Level 1: Selfie Verified
- Selfie verification approved
- Basic verification badge awarded
- Increased profile visibility
- Access to premium features

### Level 2: Document Verified
- Both selfie and document verification approved
- Advanced verification badge awarded
- Maximum profile visibility
- Full premium feature access
- Higher trust score in matching algorithm

## Error Handling

### Common Error Codes

**4xx Client Errors:**
- `400 BAD_REQUEST` - Invalid input parameters
- `401 UNAUTHORIZED` - Invalid or missing JWT token
- `403 FORBIDDEN` - Insufficient permissions
- `404 NOT_FOUND` - Verification not found
- `413 PAYLOAD_TOO_LARGE` - File size exceeds limits
- `429 TOO_MANY_REQUESTS` - Rate limit exceeded

**5xx Server Errors:**
- `500 INTERNAL_SERVER_ERROR` - Unexpected server error
- `503 SERVICE_UNAVAILABLE` - AI service temporarily unavailable

### Error Response Format

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input parameters",
    "details": {
      "field": "verification_id",
      "issue": "invalid UUID format"
    }
  }
}
```

## Webhook Integration

### Verification Status Updates

**Webhook Events:**
- `verification.approved` - Verification approved
- `verification.rejected` - Verification rejected
- `verification.expired` - Verification expired

**Webhook Payload:**
```json
{
  "event": "verification.approved",
  "data": {
    "verification_id": "550e8400-e29b-41d4-a716-446655440000",
    "user_id": "550e8400-e29b-41d4-a716-446655440001",
    "verification_type": "selfie",
    "new_level": 1,
    "timestamp": "2025-12-01T12:30:00Z"
  }
}
```

## Testing

### Integration Tests

Comprehensive test suite covering:
- Complete selfie verification flow
- Complete document verification flow
- Admin approval and rejection workflows
- Rate limiting functionality
- Input validation
- Security headers
- AI service integration
- Error handling

### Test Data

Mock services provide:
- Realistic AI analysis results
- Various confidence scores
- Different document types
- Error scenarios
- Edge cases

## Monitoring and Analytics

### Key Metrics

**Verification Statistics:**
- Total verification attempts
- Success rate by type
- Average processing time
- Rejection reasons distribution
- AI confidence score distribution

**Performance Metrics:**
- API response times
- AI service processing times
- Storage upload/download times
- Database query performance

**Security Metrics:**
- Rate limit violations
- Fraud detection alerts
- Suspicious IP addresses
- Failed authentication attempts

## Configuration

### AI Service Settings

```yaml
verification:
  ai_service:
    provider: "aws"
    region: "us-east-1"
    similarity_threshold: 0.85
    enabled: true
```

### Thresholds

```yaml
verification:
  thresholds:
    selfie_similarity_threshold: 0.85
    document_confidence_threshold: 0.80
    liveness_confidence_threshold: 0.90
    nsfw_threshold: 0.70
    manual_review_threshold: 0.75
```

### Limits

```yaml
verification:
  limits:
    max_attempts_per_day: 3
    max_attempts_per_month: 10
    cooldown_period: "24h"
    verification_expiry: "8760h"  # 365 days
    document_expiry: "17520h"  # 730 days
    max_selfie_file_size: 5242880  # 5MB
    max_document_file_size: 10485760  # 10MB
```

This comprehensive verification system provides robust, AI-powered user verification with multiple security layers, fraud detection, and efficient admin review capabilities.
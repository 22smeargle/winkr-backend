# Payment Flows Documentation

This document describes the various payment flows in the Winkr application, including subscription management, payment method handling, and error scenarios.

## Overview

The Winkr payment system supports multiple subscription tiers with different features, payment method management, and comprehensive webhook processing for real-time synchronization with Stripe.

## Subscription Plans

### Free Tier
- **Price**: $0/month
- **Features**:
  - Limited swipes per day (10)
  - Basic profile features
  - Standard messaging
  - Basic matching algorithm

### Premium Tier
- **Price**: $9.99/month
- **Features**:
  - Unlimited swipes
  - Advanced profile customization
  - Read receipts for messages
  - See who liked your profile
  - 5 Super Likes per day
  - 1 Boost per month
  - Passport feature (swipe in different locations)

### Platinum Tier
- **Price**: $19.99/month
- **Features**:
  - All Premium features
  - Unlimited Super Likes
  - 5 Boosts per month
  - Top Profile placement once per week
  - Message before match
  - Priority customer support
  - Advanced filters
  - Profile analytics

## Payment Flows

### 1. New User Subscription Flow

#### Step 1: View Available Plans
```http
GET /v1/plans
Authorization: Bearer {jwt_token}
```

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "price_free_tier",
      "name": "Free",
      "description": "Basic features for casual dating",
      "amount": 0,
      "currency": "usd",
      "interval": "month",
      "features": [
        "10 swipes per day",
        "Basic profile",
        "Standard messaging"
      ],
      "isActive": true
    },
    {
      "id": "price_premium_monthly",
      "name": "Premium",
      "description": "Enhanced features for serious dating",
      "amount": 999,
      "currency": "usd",
      "interval": "month",
      "trialPeriodDays": 7,
      "features": [
        "Unlimited swipes",
        "Advanced profile",
        "Read receipts",
        "See who liked you",
        "5 Super Likes per day",
        "1 Boost per month"
      ],
      "isActive": true
    },
    {
      "id": "price_platinum_monthly",
      "name": "Platinum",
      "description": "All features for power users",
      "amount": 1999,
      "currency": "usd",
      "interval": "month",
      "trialPeriodDays": 7,
      "features": [
        "All Premium features",
        "Unlimited Super Likes",
        "5 Boosts per month",
        "Top Profile placement",
        "Message before match",
        "Priority support"
      ],
      "isActive": true
    }
  ]
}
```

#### Step 2: Add Payment Method
```http
POST /v1/payment/methods
Authorization: Bearer {jwt_token}
Content-Type: application/json

{
  "paymentMethodId": "pm_1O2x3a2eZvKYlo2C5Zl3xY2a",
  "isDefault": true
}
```

#### Step 3: Create Subscription
```http
POST /v1/subscribe
Authorization: Bearer {jwt_token}
Content-Type: application/json

{
  "planId": "price_premium_monthly",
  "paymentMethodId": "pm_1O2x3a2eZvKYlo2C5Zl3xY2a",
  "trialPeriodDays": 7
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "sub_1O2x3a2eZvKYlo2C5Zl3xY2a",
    "userId": "550e8400-e29b-41d4-a716-446655440000",
    "planId": "price_premium_monthly",
    "planName": "Premium",
    "status": "trialing",
    "currentPeriodStart": "2025-01-01T00:00:00Z",
    "currentPeriodEnd": "2025-01-08T00:00:00Z",
    "trialStart": "2025-01-01T00:00:00Z",
    "trialEnd": "2025-01-08T00:00:00Z",
    "cancelAtPeriodEnd": false,
    "stripeSubscriptionId": "sub_1O2x3a2eZvKYlo2C5Zl3xY2a",
    "stripeCustomerId": "cus_1O2x3a2eZvKYlo2C5Zl3xY2a",
    "createdAt": "2025-01-01T00:00:00Z",
    "updatedAt": "2025-01-01T00:00:00Z"
  }
}
```

### 2. Subscription Upgrade Flow

#### Step 1: Check Current Subscription
```http
GET /v1/me/subscription
Authorization: Bearer {jwt_token}
```

#### Step 2: Update Subscription
```http
PUT /v1/subscription/update
Authorization: Bearer {jwt_token}
Content-Type: application/json

{
  "planId": "price_platinum_monthly",
  "prorationBehavior": "create_prorations"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "sub_1O2x3a2eZvKYlo2C5Zl3xY2a",
    "userId": "550e8400-e29b-41d4-a716-446655440000",
    "planId": "price_platinum_monthly",
    "planName": "Platinum",
    "status": "active",
    "currentPeriodStart": "2025-01-01T00:00:00Z",
    "currentPeriodEnd": "2025-02-01T00:00:00Z",
    "cancelAtPeriodEnd": false,
    "stripeSubscriptionId": "sub_1O2x3a2eZvKYlo2C5Zl3xY2a",
    "stripeCustomerId": "cus_1O2x3a2eZvKYlo2C5Zl3xY2a",
    "createdAt": "2025-01-01T00:00:00Z",
    "updatedAt": "2025-01-15T00:00:00Z"
  }
}
```

### 3. Subscription Cancellation Flow

#### Step 1: Cancel Subscription
```http
POST /v1/subscription/cancel
Authorization: Bearer {jwt_token}
Content-Type: application/json

{
  "cancelAtPeriodEnd": true,
  "reason": "Found a match"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "sub_1O2x3a2eZvKYlo2C5Zl3xY2a",
    "userId": "550e8400-e29b-41d4-a716-446655440000",
    "planId": "price_premium_monthly",
    "planName": "Premium",
    "status": "active",
    "currentPeriodStart": "2025-01-01T00:00:00Z",
    "currentPeriodEnd": "2025-02-01T00:00:00Z",
    "cancelAtPeriodEnd": true,
    "canceledAt": "2025-01-15T00:00:00Z",
    "stripeSubscriptionId": "sub_1O2x3a2eZvKYlo2C5Zl3xY2a",
    "stripeCustomerId": "cus_1O2x3a2eZvKYlo2C5Zl3xY2a",
    "createdAt": "2025-01-01T00:00:00Z",
    "updatedAt": "2025-01-15T00:00:00Z"
  }
}
```

#### Step 2: Reactivate Subscription (Optional)
```http
POST /v1/subscription/reactivate
Authorization: Bearer {jwt_token}
```

### 4. Payment Method Management Flow

#### Add New Payment Method
```http
POST /v1/payment/methods
Authorization: Bearer {jwt_token}
Content-Type: application/json

{
  "paymentMethodId": "pm_1O2x3a2eZvKYlo2C5Zl3xY2a",
  "isDefault": false
}
```

#### List Payment Methods
```http
GET /v1/payment/methods
Authorization: Bearer {jwt_token}
```

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "pm_1O2x3a2eZvKYlo2C5Zl3xY2a",
      "userId": "550e8400-e29b-41d4-a716-446655440000",
      "stripePaymentMethodId": "pm_1O2x3a2eZvKYlo2C5Zl3xY2a",
      "type": "card",
      "brand": "visa",
      "last4": "4242",
      "expMonth": 12,
      "expYear": 2025,
      "country": "US",
      "isDefault": true,
      "createdAt": "2025-01-01T00:00:00Z",
      "updatedAt": "2025-01-01T00:00:00Z"
    }
  ]
}
```

#### Set Default Payment Method
```http
PUT /v1/payment/methods/pm_1O2x3a2eZvKYlo2C5Zl3xY2a/default
Authorization: Bearer {jwt_token}
```

#### Remove Payment Method
```http
DELETE /v1/payment/methods/pm_1O2x3a2eZvKYlo2C5Zl3xY2a
Authorization: Bearer {jwt_token}
```

### 5. Payment History Flow

#### Get Payment History
```http
GET /v1/payment/history?limit=20&offset=0&status=succeeded
Authorization: Bearer {jwt_token}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "payments": [
      {
        "id": "pay_1O2x3a2eZvKYlo2C5Zl3xY2a",
        "userId": "550e8400-e29b-41d4-a716-446655440000",
        "stripePaymentIntentId": "pi_1O2x3a2eZvKYlo2C5Zl3xY2a",
        "amount": 999,
        "currency": "usd",
        "status": "succeeded",
        "paymentMethodId": "pm_1O2x3a2eZvKYlo2C5Zl3xY2a",
        "description": "Premium subscription",
        "createdAt": "2025-01-01T00:00:00Z",
        "updatedAt": "2025-01-01T00:00:00Z"
      }
    ],
    "total": 15,
    "hasMore": true
  }
}
```

### 6. Invoice Management Flow

#### Get Invoices
```http
GET /v1/payment/invoices?limit=20&offset=0&status=paid
Authorization: Bearer {jwt_token}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "invoices": [
      {
        "id": "in_1O2x3a2eZvKYlo2C5Zl3xY2a",
        "userId": "550e8400-e29b-41d4-a716-446655440000",
        "stripeInvoiceId": "in_1O2x3a2eZvKYlo2C5Zl3xY2a",
        "subscriptionId": "sub_1O2x3a2eZvKYlo2C5Zl3xY2a",
        "status": "paid",
        "amount": 999,
        "currency": "usd",
        "dueDate": "2025-02-01T00:00:00Z",
        "paidAt": "2025-01-01T00:00:00Z",
        "hostedInvoiceUrl": "https://invoice.stripe.com/invoice/acct_1O2x3a2eZvKYlo2C5Zl3xY2a",
        "invoicePdf": "https://invoice.stripe.com/invoice/acct_1O2x3a2eZvKYlo2C5Zl3xY2a/pdf",
        "createdAt": "2025-01-01T00:00:00Z",
        "updatedAt": "2025-01-01T00:00:00Z"
      }
    ],
    "total": 12,
    "hasMore": true
  }
}
```

#### Get Invoice Details
```http
GET /v1/payment/invoices/in_1O2x3a2eZvKYlo2C5Zl3xY2a
Authorization: Bearer {jwt_token}
```

## Error Handling

### Common Error Responses

#### Validation Error (400)
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request parameters",
    "details": [
      {
        "field": "planId",
        "message": "Plan ID is required"
      }
    ]
  }
}
```

#### Unauthorized Error (401)
```json
{
  "success": false,
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Authentication required"
  }
}
```

#### Payment Required Error (402)
```json
{
  "success": false,
  "error": {
    "code": "PAYMENT_REQUIRED",
    "message": "Payment method is required"
  }
}
```

#### Conflict Error (409)
```json
{
  "success": false,
  "error": {
    "code": "ACTIVE_SUBSCRIPTION_EXISTS",
    "message": "User already has an active subscription"
  }
}
```

#### Not Found Error (404)
```json
{
  "success": false,
  "error": {
    "code": "SUBSCRIPTION_NOT_FOUND",
    "message": "No active subscription found"
  }
}
```

#### Internal Server Error (500)
```json
{
  "success": false,
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "An unexpected error occurred"
  }
}
```

### Error Scenarios and Recovery

#### Payment Method Declined
1. **Scenario**: User tries to subscribe with a declined payment method
2. **Response**: 402 Payment Required with specific decline reason
3. **Recovery**: User should add a different payment method

#### Insufficient Funds
1. **Scenario**: Payment fails due to insufficient funds
2. **Response**: 402 Payment Required with insufficient funds message
3. **Recovery**: User should use a different payment method

#### Subscription Already Active
1. **Scenario**: User tries to create a new subscription while already having one
2. **Response**: 409 Conflict with active subscription message
3. **Recovery**: User should update existing subscription instead

#### Plan Not Found
1. **Scenario**: User tries to subscribe to a non-existent plan
2. **Response**: 400 Validation Error with plan not found message
3. **Recovery**: User should select from available plans

#### Webhook Processing Failure
1. **Scenario**: Stripe webhook fails to process
2. **Response**: Logged error, automatic retry
3. **Recovery**: Manual intervention may be required

## Security Considerations

### Payment Security
- All payment operations require authentication
- Payment method IDs are never stored in plain text
- Webhook signatures are verified before processing
- Sensitive payment data is handled by Stripe only

### Rate Limiting
- Payment endpoints have stricter rate limits
- Webhook endpoint has special handling without authentication
- Failed payment attempts are tracked and limited

### Data Protection
- Payment history is only accessible to the user
- Payment methods are isolated per user
- Audit logs are maintained for all payment operations

## Testing Payment Flows

### Using Stripe Test Cards
Use Stripe test cards to test various payment scenarios:

- **Success**: 4242 4242 4242 4242
- **Declined**: 4000 0000 0000 0002
- **Insufficient Funds**: 4000 0000 0000 9995
- **Expired Card**: 4000 0000 0000 0069

### Testing Webhooks
Use Stripe CLI to test webhook events:
```bash
stripe listen --forward-to localhost:8080/v1/payment/webhook
stripe trigger customer.subscription.created
stripe trigger invoice.payment_succeeded
stripe trigger invoice.payment_failed
```

### Testing Subscription Changes
1. Create a test subscription
2. Update to different plans
3. Test cancellation and reactivation
4. Verify webhook processing

## Best Practices

### For Developers
1. Always handle payment errors gracefully
2. Implement proper retry logic for failed payments
3. Use webhooks to keep data synchronized
4. Cache subscription status for performance
5. Log all payment operations for audit

### For Users
1. Keep payment methods up to date
2. Monitor subscription status and billing dates
3. Review invoices regularly
4. Contact support for payment issues
5. Use secure payment methods

## Monitoring and Analytics

### Key Metrics
- Subscription conversion rate
- Payment success rate
- Churn rate by plan
- Average revenue per user
- Payment method distribution

### Alerts
- Payment failure rate above 5%
- Webhook processing failures
- Subscription cancellations spike
- Revenue anomalies

## Support and Troubleshooting

### Common Issues
1. **Payment not processing**: Check payment method validity
2. **Subscription not activating**: Verify webhook processing
3. **Features not unlocked**: Check subscription status
4. **Billing issues**: Review payment history

### Support Process
1. Check user's subscription status
2. Review payment history and invoices
3. Verify webhook event processing
4. Check for system-wide issues
5. Escalate to Stripe support if needed
# Stripe Webhook Events Documentation

This document describes all Stripe webhook events that the Winkr application processes and their corresponding payloads.

## Overview

The Winkr application processes various Stripe webhook events to keep the local database synchronized with Stripe's state. All webhook events are received at the `/payment/webhook` endpoint and are verified using the Stripe signature.

## Webhook Event Types

### 1. Customer Events

#### customer.created
Triggered when a new customer is created in Stripe.

**Payload Structure:**
```json
{
  "id": "evt_1O2x3a2eZvKYlo2C5Zl3xY2a",
  "object": "event",
  "api_version": "2025-10-16",
  "created": 1672531200,
  "type": "customer.created",
  "data": {
    "object": {
      "id": "cus_1O2x3a2eZvKYlo2C5Zl3xY2a",
      "object": "customer",
      "email": "user@example.com",
      "name": "John Doe",
      "metadata": {
        "user_id": "550e8400-e29b-41d4-a716-446655440000"
      }
    }
  }
}
```

**Processing Logic:**
- Create a new customer record in the local database
- Associate the customer with the user based on metadata
- Cache the customer information

#### customer.updated
Triggered when a customer's information is updated in Stripe.

**Processing Logic:**
- Update the local customer record
- Invalidate cached customer information
- Log the update for audit purposes

#### customer.deleted
Triggered when a customer is deleted in Stripe.

**Processing Logic:**
- Mark the local customer record as deleted
- Cancel all active subscriptions
- Invalidate all cached customer data

### 2. Subscription Events

#### invoice.payment_succeeded
Triggered when an invoice payment is successful.

**Payload Structure:**
```json
{
  "id": "evt_1O2x3a2eZvKYlo2C5Zl3xY2a",
  "object": "event",
  "type": "invoice.payment_succeeded",
  "data": {
    "object": {
      "id": "in_1O2x3a2eZvKYlo2C5Zl3xY2a",
      "object": "invoice",
      "customer": "cus_1O2x3a2eZvKYlo2C5Zl3xY2a",
      "subscription": "sub_1O2x3a2eZvKYlo2C5Zl3xY2a",
      "status": "paid",
      "amount_paid": 999,
      "currency": "usd"
    }
  }
}
```

**Processing Logic:**
- Update subscription status to active
- Record successful payment
- Update user's subscription features
- Send confirmation notification

#### invoice.payment_failed
Triggered when an invoice payment fails.

**Processing Logic:**
- Update subscription status to past_due
- Record failed payment
- Send payment failure notification
- Schedule retry attempts

#### customer.subscription.created
Triggered when a new subscription is created.

**Payload Structure:**
```json
{
  "id": "evt_1O2x3a2eZvKYlo2C5Zl3xY2a",
  "object": "event",
  "type": "customer.subscription.created",
  "data": {
    "object": {
      "id": "sub_1O2x3a2eZvKYlo2C5Zl3xY2a",
      "object": "subscription",
      "customer": "cus_1O2x3a2eZvKYlo2C5Zl3xY2a",
      "status": "active",
      "current_period_start": 1672531200,
      "current_period_end": 1675209600,
      "items": {
        "data": [
          {
            "price": {
              "id": "price_1O2x3a2eZvKYlo2C5Zl3xY2a",
              "product": "prod_1O2x3a2eZvKYlo2C5Zl3xY2a"
            }
          }
        ]
      }
    }
  }
}
```

**Processing Logic:**
- Create subscription record in local database
- Update user's subscription status
- Grant subscription features to user
- Cache subscription information

#### customer.subscription.updated
Triggered when a subscription is updated.

**Processing Logic:**
- Update local subscription record
- Handle plan changes (upgrade/downgrade)
- Update user's features based on new plan
- Invalidate cached subscription data

#### customer.subscription.deleted
Triggered when a subscription is canceled/deleted.

**Processing Logic:**
- Mark subscription as canceled
- Remove subscription features from user
- Handle grace period if applicable
- Invalidate cached subscription data

#### customer.subscription.trial_will_end
Triggered three days before a trial ends.

**Processing Logic:**
- Send trial ending notification
- Offer upgrade options
- Prepare for subscription activation

### 3. Payment Method Events

#### payment_method.attached
Triggered when a payment method is attached to a customer.

**Payload Structure:**
```json
{
  "id": "evt_1O2x3a2eZvKYlo2C5Zl3xY2a",
  "object": "event",
  "type": "payment_method.attached",
  "data": {
    "object": {
      "id": "pm_1O2x3a2eZvKYlo2C5Zl3xY2a",
      "object": "payment_method",
      "customer": "cus_1O2x3a2eZvKYlo2C5Zl3xY2a",
      "type": "card",
      "card": {
        "brand": "visa",
        "last4": "4242",
        "exp_month": 12,
        "exp_year": 2025
      }
    }
  }
}
```

**Processing Logic:**
- Create payment method record in local database
- Associate with user based on customer
- Cache payment method information

#### payment_method.detached
Triggered when a payment method is detached from a customer.

**Processing Logic:**
- Mark payment method as deleted in local database
- Update default payment method if necessary
- Invalidate cached payment method data

#### payment_method.updated
Triggered when a payment method is updated.

**Processing Logic:**
- Update local payment method record
- Invalidate cached payment method data

### 4. Payment Intent Events

#### payment_intent.succeeded
Triggered when a payment intent is successfully completed.

**Payload Structure:**
```json
{
  "id": "evt_1O2x3a2eZvKYlo2C5Zl3xY2a",
  "object": "event",
  "type": "payment_intent.succeeded",
  "data": {
    "object": {
      "id": "pi_1O2x3a2eZvKYlo2C5Zl3xY2a",
      "object": "payment_intent",
      "customer": "cus_1O2x3a2eZvKYlo2C5Zl3xY2a",
      "amount": 999,
      "currency": "usd",
      "status": "succeeded",
      "payment_method": "pm_1O2x3a2eZvKYlo2C5Zl3xY2a"
    }
  }
}
```

**Processing Logic:**
- Create payment record in local database
- Update associated invoice status
- Send payment confirmation
- Cache payment information

#### payment_intent.payment_failed
Triggered when a payment intent fails.

**Processing Logic:**
- Create failed payment record
- Update invoice status
- Send payment failure notification
- Log failure for analysis

### 5. Invoice Events

#### invoice.created
Triggered when a new invoice is created.

**Processing Logic:**
- Create invoice record in local database
- Cache invoice information
- Prepare for payment processing

#### invoice.finalized
Triggered when an invoice is finalized.

**Processing Logic:**
- Update invoice status
- Send invoice notification
- Prepare for payment collection

#### invoice.voided
Triggered when an invoice is voided.

**Processing Logic:**
- Update invoice status to void
- Handle any associated refunds
- Update subscription if necessary

### 6. Refund Events

#### charge.dispute.created
Triggered when a charge dispute is created.

**Processing Logic:**
- Create dispute record
- Notify admin team
- Update payment status
- Prepare dispute response

#### refund.created
Triggered when a refund is created.

**Payload Structure:**
```json
{
  "id": "evt_1O2x3a2eZvKYlo2C5Zl3xY2a",
  "object": "event",
  "type": "refund.created",
  "data": {
    "object": {
      "id": "re_1O2x3a2eZvKYlo2C5Zl3xY2a",
      "object": "refund",
      "charge": "ch_1O2x3a2eZvKYlo2C5Zl3xY2a",
      "amount": 999,
      "currency": "usd",
      "status": "succeeded",
      "reason": "requested_by_customer"
    }
  }
}
```

**Processing Logic:**
- Create refund record in local database
- Update payment status
- Send refund confirmation
- Cache refund information

#### refund.updated
Triggered when a refund is updated.

**Processing Logic:**
- Update local refund record
- Handle status changes
- Invalidate cached refund data

## Webhook Processing Flow

### 1. Signature Verification
All incoming webhooks are verified using the Stripe signature:
```go
func (s *StripeService) VerifyWebhookSignature(payload []byte, signature string) error {
    // Verify webhook signature using Stripe's webhook secret
}
```

### 2. Event Processing
Each webhook event is processed through the following steps:
1. Parse the event JSON
2. Verify the event type
3. Extract relevant data
4. Update local database
5. Invalidate cache as needed
6. Send notifications
7. Log the event for audit

### 3. Error Handling
- Webhook events are retried on failure
- Failed events are logged for manual review
- Critical errors trigger admin notifications

### 4. Idempotency
All webhook processing is idempotent:
- Events are deduplicated using event ID
- Processing state is tracked in the database
- Failed events can be safely retried

## Webhook Configuration

### Endpoint URL
```
https://api.winkr.com/v1/payment/webhook
```

### Events to Subscribe To
- customer.created
- customer.updated
- customer.deleted
- invoice.payment_succeeded
- invoice.payment_failed
- customer.subscription.created
- customer.subscription.updated
- customer.subscription.deleted
- customer.subscription.trial_will_end
- payment_method.attached
- payment_method.detached
- payment_method.updated
- payment_intent.succeeded
- payment_intent.payment_failed
- invoice.created
- invoice.finalized
- invoice.voided
- charge.dispute.created
- refund.created
- refund.updated

### Security Considerations
- All webhooks must be signed by Stripe
- Webhook endpoint is rate limited
- Failed webhook attempts are logged
- Webhook processing is monitored for anomalies

## Testing Webhooks

### Using Stripe CLI
```bash
# Forward webhook events to local development server
stripe listen --forward-to localhost:8080/v1/payment/webhook

# Trigger specific events for testing
stripe trigger customer.subscription.created
stripe trigger invoice.payment_succeeded
stripe trigger payment_intent.succeeded
```

### Using Stripe Dashboard
1. Go to Developers > Webhooks
2. Select your webhook endpoint
3. Click "Send test webhook" to send test events

## Monitoring and Alerting

### Webhook Health Monitoring
- Track webhook delivery success rate
- Monitor webhook processing time
- Alert on webhook failures
- Track retry attempts

### Key Metrics
- Webhook delivery success rate
- Average processing time
- Failed webhook events
- Retry attempts per event

### Alerting Rules
- Alert when webhook success rate drops below 95%
- Alert when webhook processing time exceeds 5 seconds
- Alert when webhook failures exceed 10 per hour
- Alert when retry queue exceeds 100 events

## Troubleshooting

### Common Issues
1. **Signature Verification Failed**
   - Check webhook secret configuration
   - Verify request body is not modified
   - Ensure timestamp is within tolerance

2. **Duplicate Event Processing**
   - Check event ID deduplication
   - Verify idempotency handling
   - Review database transaction handling

3. **Missing Events**
   - Verify webhook endpoint is registered
   - Check event subscription list
   - Review webhook delivery logs

4. **Processing Errors**
   - Check database connectivity
   - Verify cache service status
   - Review error logs for details

### Debugging Steps
1. Check webhook signature verification
2. Review event parsing logic
3. Verify database operations
4. Check cache invalidation
5. Review notification sending
6. Analyze error logs

## Best Practices

1. **Always verify webhook signatures** before processing events
2. **Make processing idempotent** to handle duplicate events
3. **Use database transactions** for data consistency
4. **Implement proper error handling** and logging
5. **Monitor webhook health** and set up alerts
6. **Test webhook processing** thoroughly
7. **Keep webhook processing fast** to avoid timeouts
8. **Use background jobs** for heavy processing
9. **Implement retry logic** for failed events
10. **Document webhook processing** for maintenance
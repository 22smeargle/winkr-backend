# WebSocket Events Documentation

This document describes the WebSocket events used for real-time chat functionality in the Winkr dating application.

## Connection

### WebSocket Endpoint
```
ws://localhost:8080/ws
wss://api.winkr.com/ws
```

### Authentication
WebSocket connections require authentication via the `Authorization` header:
```
Authorization: Bearer <jwt_token>
```

### Connection Flow
1. Client establishes WebSocket connection with JWT token
2. Server validates token and authenticates connection
3. Server sends `connection:established` event
4. Client can now send and receive events

## Event Format

All WebSocket messages follow this JSON format:
```json
{
  "event": "event_name",
  "data": {
    // Event-specific data
  },
  "timestamp": "2025-01-01T12:00:00Z"
}
```

## Client-to-Server Events

### message:send
Send a new message to a conversation.

```json
{
  "event": "message:send",
  "data": {
    "conversation_id": "conv-uuid-1",
    "content": "Hey! How are you?",
    "type": "text",
    "metadata": {}
  }
}
```

**Response Events:**
- `message:new` - Message sent successfully
- `error` - Failed to send message

### message:read
Mark messages as read in a conversation.

```json
{
  "event": "message:read",
  "data": {
    "conversation_id": "conv-uuid-1",
    "message_id": "msg-uuid-1"
  }
}
```

**Response Events:**
- `message:viewed` - Messages marked as read
- `error` - Failed to mark messages as read

### message:delete
Delete a specific message.

```json
{
  "event": "message:delete",
  "data": {
    "conversation_id": "conv-uuid-1",
    "message_id": "msg-uuid-1"
  }
}
```

**Response Events:**
- `message:deleted` - Message deleted successfully
- `error` - Failed to delete message

### typing:start
Indicate that user started typing in a conversation.

```json
{
  "event": "typing:start",
  "data": {
    "conversation_id": "conv-uuid-1"
  }
}
```

**Response Events:**
- `typing:indicator` - Typing indicator broadcast to other user
- `error` - Failed to send typing indicator

### typing:stop
Indicate that user stopped typing in a conversation.

```json
{
  "event": "typing:stop",
  "data": {
    "conversation_id": "conv-uuid-1"
  }
}
```

**Response Events:**
- `typing:indicator` - Typing indicator cleared for other user
- `error` - Failed to clear typing indicator

### conversation:join
Join a conversation room to receive real-time updates.

```json
{
  "event": "conversation:join",
  "data": {
    "conversation_id": "conv-uuid-1"
  }
}
```

**Response Events:**
- `conversation:joined` - Successfully joined conversation
- `error` - Failed to join conversation

### conversation:leave
Leave a conversation room to stop receiving updates.

```json
{
  "event": "conversation:leave",
  "data": {
    "conversation_id": "conv-uuid-1"
  }
}
```

**Response Events:**
- `conversation:left` - Successfully left conversation
- `error` - Failed to leave conversation

### user:status
Update user online status.

```json
{
  "event": "user:status",
  "data": {
    "status": "online"
  }
}
```

**Response Events:**
- `user:status_updated` - Status updated successfully
- `error` - Failed to update status

## Server-to-Client Events

### connection:established
Sent when WebSocket connection is successfully established.

```json
{
  "event": "connection:established",
  "data": {
    "user_id": "user-uuid-1",
    "connection_id": "conn-uuid-1",
    "timestamp": "2025-01-01T12:00:00Z"
  }
}
```

### message:new
Sent when a new message is received in a conversation.

```json
{
  "event": "message:new",
  "data": {
    "message": {
      "id": "msg-uuid-1",
      "conversation_id": "conv-uuid-1",
      "sender_id": "user-uuid-2",
      "content": "Hey! How are you?",
      "type": "text",
      "created_at": "2025-01-01T12:00:00Z",
      "updated_at": "2025-01-01T12:00:00Z",
      "is_read": false,
      "metadata": {}
    },
    "sender": {
      "id": "user-uuid-2",
      "username": "jane_doe",
      "avatar_url": "https://example.com/avatar.jpg"
    }
  }
}
```

### message:delivered
Sent when a message is delivered to the recipient.

```json
{
  "event": "message:delivered",
  "data": {
    "message_id": "msg-uuid-1",
    "conversation_id": "conv-uuid-1",
    "delivered_at": "2025-01-01T12:00:05Z"
  }
}
```

### message:viewed
Sent when messages are marked as read by the recipient.

```json
{
  "event": "message:viewed",
  "data": {
    "message_id": "msg-uuid-1",
    "conversation_id": "conv-uuid-1",
    "viewed_at": "2025-01-01T12:00:10Z"
  }
}
```

### message:deleted
Sent when a message is deleted by either participant.

```json
{
  "event": "message:deleted",
  "data": {
    "message_id": "msg-uuid-1",
    "conversation_id": "conv-uuid-1",
    "deleted_by": "user-uuid-1",
    "deleted_at": "2025-01-01T12:00:15Z"
  }
}
```

### typing:indicator
Sent when a user starts or stops typing in a conversation.

```json
{
  "event": "typing:indicator",
  "data": {
    "conversation_id": "conv-uuid-1",
    "user_id": "user-uuid-2",
    "username": "jane_doe",
    "is_typing": true,
    "timestamp": "2025-01-01T12:00:00Z"
  }
}
```

### user:online
Sent when a user comes online.

```json
{
  "event": "user:online",
  "data": {
    "user_id": "user-uuid-2",
    "username": "jane_doe",
    "online_at": "2025-01-01T12:00:00Z"
  }
}
```

### user:offline
Sent when a user goes offline.

```json
{
  "event": "user:offline",
  "data": {
    "user_id": "user-uuid-2",
    "username": "jane_doe",
    "offline_at": "2025-01-01T12:30:00Z"
  }
}
```

### conversation:unread
Sent when there are unread messages in a conversation.

```json
{
  "event": "conversation:unread",
  "data": {
    "conversation_id": "conv-uuid-1",
    "unread_count": 3,
    "last_message": {
      "id": "msg-uuid-1",
      "content": "Hey! How are you?",
      "type": "text",
      "sender_id": "user-uuid-2",
      "created_at": "2025-01-01T12:00:00Z"
    }
  }
}
```

### conversation:joined
Sent when successfully joined a conversation room.

```json
{
  "event": "conversation:joined",
  "data": {
    "conversation_id": "conv-uuid-1",
    "joined_at": "2025-01-01T12:00:00Z"
  }
}
```

### conversation:left
Sent when successfully left a conversation room.

```json
{
  "event": "conversation:left",
  "data": {
    "conversation_id": "conv-uuid-1",
    "left_at": "2025-01-01T12:00:00Z"
  }
}
```

### error
Sent when an error occurs with a client request.

```json
{
  "event": "error",
  "data": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid message content",
    "details": {
      "field": "content",
      "error": "Message content cannot be empty"
    },
    "request_event": "message:send"
  }
}
```

### connection:closed
Sent when the WebSocket connection is closed.

```json
{
  "event": "connection:closed",
  "data": {
    "reason": "normal_closure",
    "code": 1000,
    "timestamp": "2025-01-01T12:30:00Z"
  }
}
```

## Error Codes

| Code | Description |
|------|-------------|
| `VALIDATION_ERROR` | Invalid request data |
| `UNAUTHORIZED` | Authentication failed |
| `FORBIDDEN` | Access to resource forbidden |
| `NOT_FOUND` | Conversation or message not found |
| `RATE_LIMITED` | Too many requests |
| `INTERNAL_ERROR` | Internal server error |
| `CONNECTION_ERROR` | WebSocket connection error |
| `MESSAGE_TOO_LARGE` | Message exceeds size limit |
| `INVALID_MESSAGE_TYPE` | Unsupported message type |
| `CONVERSATION_FULL` | Cannot join conversation |

## Rate Limits

- **Messages**: 30 messages per minute, 500 per hour, 2000 per day
- **Typing Indicators**: 20 per minute
- **Connections**: 10 connections per minute
- **Conversations**: 50 new conversations per day

## Message Types

| Type | Description | Max Size |
|------|-------------|-----------|
| `text` | Plain text messages | 2000 characters |
| `photo` | Photo messages with URL | 10MB |
| `photo_ephemeral` | Self-destructing photos | 10MB |
| `location` | Location sharing | N/A |
| `system` | System notifications | N/A |
| `gift` | Virtual gifts | N/A |

## Connection Limits

- **Max connections per user**: 5
- **Connection timeout**: 30 seconds
- **Ping interval**: 30 seconds
- **Pong wait**: 60 seconds
- **Max message size**: 32KB

## Security Considerations

1. **Authentication**: All connections require valid JWT token
2. **Authorization**: Users can only access their own conversations
3. **Content Filtering**: Messages are filtered for inappropriate content
4. **Rate Limiting**: Requests are rate-limited per user
5. **Encryption**: Messages can be encrypted end-to-end
6. **PII Protection**: Personal information is detected and protected

## Best Practices

1. **Connection Management**: 
   - Close connections when app goes to background
   - Reconnect with exponential backoff
   - Handle connection errors gracefully

2. **Message Handling**:
   - Validate message content before sending
   - Handle different message types appropriately
   - Implement local caching for better UX

3. **Typing Indicators**:
   - Send `typing:start` only when user actually starts typing
   - Send `typing:stop` after 3 seconds of inactivity
   - Debounce typing events to avoid spam

4. **Error Handling**:
   - Implement proper error handling for all events
   - Show user-friendly error messages
   - Log errors for debugging

5. **Performance**:
   - Use pagination for message history
   - Implement local caching
   - Optimize image loading for photo messages
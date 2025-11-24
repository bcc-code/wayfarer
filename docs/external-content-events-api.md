# External Content Events API

This document describes the webhook API for external systems to report content completion events (reading or watching) to the Wayfarer backend.

## Overview

The External Content Events API allows external systems to notify Wayfarer when a user has finished reading or watching content. Events are stored for further processing and can be linked to achievements or other gamification features.

## Authentication

All requests must include an API key in the Authorization header using the Bearer token format.

### Header Format

```
Authorization: Bearer <your-api-key>
```

### Obtaining an API Key

Contact the Wayfarer system administrator to obtain an API key. Each external system will be assigned a unique API key with a source identifier.

## Endpoint

### Submit Content Event

**POST** `/api/v1/content-events`

Submit a new content completion event.

#### Request Headers

| Header | Value | Required |
|--------|-------|----------|
| Authorization | Bearer <api-key> | Yes |
| Content-Type | application/json | Yes |

#### Request Body

```json
{
  "person_id": "550e8400-e29b-41d4-a716-446655440000",
  "content_id": "article-123",
  "reading_plan_id": "plan-456",
  "content_progress": 0.75
}
```

**Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| person_id | string (UUID) | Yes | The Brunstad TV person_id (UUID format) |
| content_id | string | Yes | External content identifier |
| reading_plan_id | string | No | External reading plan identifier (optional) |
| content_progress | float | No | Content completion progress (0.01 to 1.1, where 1.0 = 100%). Values outside this range are ignored and stored as NULL. |

#### Success Response

**Status Code:** 201 Accepted

No response body is returned.

#### Error Responses

**400 Bad Request**

Invalid request body or parameters.

```json
{
  "error": "invalid request body",
  "details": "person_id is required"
}
```

**401 Unauthorized**

Missing or invalid API key.

```json
{
  "error": "Authorization header required"
}
```

or

```json
{
  "error": "Invalid API key"
}
```

**500 Internal Server Error**

Server error while processing the request.

```json
{
  "error": "failed to create event"
}
```
## Support

For API key requests or technical support, contact the Wayfarer system administrator.

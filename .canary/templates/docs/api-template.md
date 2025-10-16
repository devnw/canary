# {{FEATURE}} API Reference

**Requirement:** {{REQ_ID}}
**API Version:** 1.0
**Last Updated:** {{DATE}}

## Overview

Brief description of the API, its purpose, and primary use cases.

## Authentication

### Method

Describe authentication method (API keys, OAuth, JWT, etc.)

```bash
# Example authentication
curl -H "Authorization: Bearer YOUR_TOKEN" \
     https://api.example.com/v1/resource
```

### Obtaining Credentials

How to get API credentials:
1. Step 1
2. Step 2
3. Step 3

## Base URL

```
Production: https://api.example.com/v1
Staging: https://api-staging.example.com/v1
Development: http://localhost:8080/v1
```

## Endpoints

### List Resources

```http
GET /resources
```

**Description:** Retrieve a list of resources with optional filtering.

**Query Parameters:**

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| page | integer | No | 1 | Page number for pagination |
| limit | integer | No | 20 | Items per page (max 100) |
| sort | string | No | "created_at" | Sort field |
| order | string | No | "desc" | Sort order (asc/desc) |
| filter | string | No | - | Filter expression |

**Example Request:**

```bash
curl -X GET "https://api.example.com/v1/resources?page=1&limit=10&sort=name" \
     -H "Authorization: Bearer YOUR_TOKEN"
```

**Example Response (200 OK):**

```json
{
  "data": [
    {
      "id": "res_123abc",
      "name": "Resource Name",
      "status": "active",
      "created_at": "2025-10-16T18:00:00Z",
      "updated_at": "2025-10-16T18:30:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 100,
    "pages": 10
  }
}
```

**Error Responses:**

- `401 Unauthorized`: Invalid or missing authentication token
- `403 Forbidden`: Insufficient permissions
- `429 Too Many Requests`: Rate limit exceeded

---

### Get Resource

```http
GET /resources/{id}
```

**Description:** Retrieve a single resource by ID.

**Path Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| id | string | Yes | Resource identifier |

**Example Request:**

```bash
curl -X GET "https://api.example.com/v1/resources/res_123abc" \
     -H "Authorization: Bearer YOUR_TOKEN"
```

**Example Response (200 OK):**

```json
{
  "id": "res_123abc",
  "name": "Resource Name",
  "description": "Detailed description",
  "status": "active",
  "metadata": {
    "key1": "value1",
    "key2": "value2"
  },
  "created_at": "2025-10-16T18:00:00Z",
  "updated_at": "2025-10-16T18:30:00Z"
}
```

**Error Responses:**

- `404 Not Found`: Resource does not exist

---

### Create Resource

```http
POST /resources
```

**Description:** Create a new resource.

**Request Headers:**

| Header | Required | Description |
|--------|----------|-------------|
| Content-Type | Yes | Must be `application/json` |

**Request Body:**

```json
{
  "name": "string (required, max 255 chars)",
  "description": "string (optional)",
  "status": "string (optional, default: 'active')",
  "metadata": {
    "key": "value"
  }
}
```

**Field Validation:**

- `name`: Required, 1-255 characters, alphanumeric and spaces
- `description`: Optional, max 1000 characters
- `status`: Optional, one of: "active", "inactive", "pending"
- `metadata`: Optional, max 10 key-value pairs

**Example Request:**

```bash
curl -X POST "https://api.example.com/v1/resources" \
     -H "Authorization: Bearer YOUR_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "name": "New Resource",
       "description": "This is a new resource",
       "status": "active"
     }'
```

**Example Response (201 Created):**

```json
{
  "id": "res_456def",
  "name": "New Resource",
  "description": "This is a new resource",
  "status": "active",
  "metadata": {},
  "created_at": "2025-10-16T18:45:00Z",
  "updated_at": "2025-10-16T18:45:00Z"
}
```

**Error Responses:**

- `400 Bad Request`: Invalid input data
  ```json
  {
    "error": "validation_error",
    "message": "Invalid request body",
    "details": [
      {
        "field": "name",
        "error": "required field missing"
      }
    ]
  }
  ```
- `409 Conflict`: Resource already exists

---

### Update Resource

```http
PUT /resources/{id}
PATCH /resources/{id}
```

**Description:** 
- `PUT`: Full replacement of resource (all fields required)
- `PATCH`: Partial update (only specified fields updated)

**Path Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| id | string | Yes | Resource identifier |

**Request Body (PATCH):**

```json
{
  "name": "string (optional)",
  "description": "string (optional)",
  "status": "string (optional)"
}
```

**Example Request:**

```bash
curl -X PATCH "https://api.example.com/v1/resources/res_123abc" \
     -H "Authorization: Bearer YOUR_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "status": "inactive"
     }'
```

**Example Response (200 OK):**

```json
{
  "id": "res_123abc",
  "name": "Resource Name",
  "description": "Detailed description",
  "status": "inactive",
  "metadata": {},
  "created_at": "2025-10-16T18:00:00Z",
  "updated_at": "2025-10-16T19:00:00Z"
}
```

**Error Responses:**

- `404 Not Found`: Resource does not exist
- `422 Unprocessable Entity`: Invalid field values

---

### Delete Resource

```http
DELETE /resources/{id}
```

**Description:** Permanently delete a resource.

**Path Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| id | string | Yes | Resource identifier |

**Example Request:**

```bash
curl -X DELETE "https://api.example.com/v1/resources/res_123abc" \
     -H "Authorization: Bearer YOUR_TOKEN"
```

**Example Response (204 No Content):**

No response body.

**Error Responses:**

- `404 Not Found`: Resource does not exist
- `409 Conflict`: Resource cannot be deleted (dependencies exist)

---

## Data Models

### Resource

```json
{
  "id": "string (UUID)",
  "name": "string",
  "description": "string (optional)",
  "status": "string (enum: active, inactive, pending)",
  "metadata": "object (key-value pairs)",
  "created_at": "string (ISO 8601 timestamp)",
  "updated_at": "string (ISO 8601 timestamp)"
}
```

**Field Descriptions:**

- `id`: Unique identifier, auto-generated
- `name`: Human-readable name
- `description`: Optional detailed description
- `status`: Current resource state
- `metadata`: Flexible key-value storage
- `created_at`: Timestamp of creation
- `updated_at`: Timestamp of last modification

## Error Handling

### Error Response Format

All errors follow this structure:

```json
{
  "error": "error_code",
  "message": "Human-readable error message",
  "details": {},
  "request_id": "req_xyz789"
}
```

### Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| validation_error | 400 | Request validation failed |
| unauthorized | 401 | Authentication required |
| forbidden | 403 | Insufficient permissions |
| not_found | 404 | Resource not found |
| conflict | 409 | Resource conflict |
| rate_limit_exceeded | 429 | Too many requests |
| internal_error | 500 | Server error |

## Rate Limiting

**Limits:**
- 100 requests per minute per API key
- 1000 requests per hour per API key

**Headers:**

```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1634400000
```

**When limit exceeded:**

```json
{
  "error": "rate_limit_exceeded",
  "message": "Rate limit exceeded. Try again in 30 seconds.",
  "retry_after": 30
}
```

## Pagination

All list endpoints support pagination:

**Request:**
```
GET /resources?page=2&limit=20
```

**Response:**
```json
{
  "data": [...],
  "pagination": {
    "page": 2,
    "limit": 20,
    "total": 100,
    "pages": 5,
    "has_next": true,
    "has_prev": true
  }
}
```

## Filtering

Support for filter expressions:

**Syntax:**
```
GET /resources?filter=status:active,name:~test
```

**Operators:**
- `:` - Equals
- `:~` - Contains
- `:>` - Greater than
- `:<` - Less than

**Examples:**
```
?filter=status:active                    # Status equals active
?filter=name:~test                       # Name contains "test"
?filter=created_at:>2025-01-01          # Created after date
?filter=status:active,name:~test        # Multiple filters (AND)
```

## Sorting

**Syntax:**
```
GET /resources?sort=name&order=asc
```

**Available sort fields:**
- name
- created_at
- updated_at
- status

**Sort order:**
- `asc` - Ascending
- `desc` - Descending (default)

## Versioning

API versioning via URL path:

```
/v1/resources  # Version 1
/v2/resources  # Version 2 (future)
```

**Version compatibility:**
- v1: Current version, stable
- Breaking changes require new major version

## SDKs & Libraries

### Official SDKs

- **Go:** `go get github.com/example/api-go`
- **Python:** `pip install example-api`
- **JavaScript:** `npm install @example/api`

### Code Examples

**Go:**
```go
client := api.NewClient("YOUR_TOKEN")
resources, err := client.ListResources(api.ListOptions{
    Page: 1,
    Limit: 10,
})
```

**Python:**
```python
from example_api import Client

client = Client("YOUR_TOKEN")
resources = client.list_resources(page=1, limit=10)
```

**JavaScript:**
```javascript
const { Client } = require('@example/api');

const client = new Client('YOUR_TOKEN');
const resources = await client.listResources({ page: 1, limit: 10 });
```

## Webhooks

### Subscribing to Events

```http
POST /webhooks
```

**Request:**
```json
{
  "url": "https://your-app.com/webhook",
  "events": ["resource.created", "resource.updated"],
  "secret": "your_webhook_secret"
}
```

### Event Payload

```json
{
  "event": "resource.created",
  "timestamp": "2025-10-16T18:00:00Z",
  "data": {
    "id": "res_123abc",
    "name": "New Resource"
  }
}
```

### Verifying Webhook Signatures

```go
signature := request.Header.Get("X-Webhook-Signature")
expected := hmac.SHA256(payload, secret)
if signature != expected {
    // Invalid signature
}
```

## Changelog

### Version 1.0 (2025-10-16)

- Initial API release
- CRUD operations for resources
- Pagination and filtering support

---

*Generated from api-template.md*

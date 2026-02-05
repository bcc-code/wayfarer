# Cryptex Admin URL Endpoint

## Overview

The Wayfarer backend provides an endpoint that generates admin login URLs for Cryptex integration. These URLs contain JWT tokens with user and church information that Cryptex can use to identify and authorize users.

## Endpoint

```
GET /plugins/ladder-to-heaven/cryptex-admin-url
```

## Authentication

The endpoint requires a valid Wayfarer JWT token in the Authorization header:

```
Authorization: Bearer <wayfarer-jwt-token>
```

### Required Roles

The authenticated user must have at least one of the following roles (lowercase):
- `church_admin`
- `admin`
- `superadmin`

Users with only the `user` role will receive a 403 Forbidden response.

## Response

### Success (200 OK)

```json
{
  "url": "https://cryptex.example.com/callback?token=<jwt-token-string>"
}
```

The URL format is: `<CRYPTEX_BASE_URL>/callback?token=<jwt>`

### Error Responses

| Status | Error | Description |
|--------|-------|-------------|
| 401 | `authentication required` | Missing or invalid Authorization header |
| 403 | `insufficient permissions` | User lacks required role |
| 500 | `failed to retrieve user data` | Database error fetching user |
| 500 | `failed to retrieve church data` | Database error fetching church |
| 500 | `failed to retrieve project configuration` | Error getting current project ID |
| 500 | `failed to generate token` | Error signing the JWT |
| 503 | `feature disabled` | `PLUGIN_LADDER_TO_HEAVEN_CRYPTEX_SECRET_KEY` or `PLUGIN_LADDER_TO_HEAVEN_CRYPTEX_BASE_URL` not configured |

## JWT Token Structure

### Header

```json
{
  "alg": "HS256",
  "typ": "JWT"
}
```

### Payload (Claims)

All custom claims use **camelCase** naming:

```json
{
  "userName": "John Doe",
  "userId": "US01ARZ3NDEKTSV4RRFFQ69G5FAV",
  "churchId": "CH01ARZ3NDEKTSV4RRFFQ69G5FAV",
  "churchName": "Oslo Church",
  "projectId": "PR01ARZ3NDEKTSV4RRFFQ69G5FAV",
  "aud": ["LADD-Cryptex"],
  "exp": 1737641234,
  "iat": 1721916034,
  "iss": "wayfarer"
}
```

### Claim Descriptions

| Claim | Type | Description |
|-------|------|-------------|
| `userName` | string | User's full name from Wayfarer |
| `userId` | string | User's Wayfarer ID (prefixed ULID, 28 chars) |
| `churchId` | string | User's church ID (prefixed ULID, 28 chars) |
| `churchName` | string | Name of the user's church |
| `projectId` | string | Current active project ID from Wayfarer settings |
| `aud` | string[] | Audience - always `["LADD-Cryptex"]` |
| `exp` | number | Expiration time (Unix timestamp) - 6 months from issuance |
| `iat` | number | Issued at time (Unix timestamp) |
| `iss` | string | Issuer - value of `JWT_ISSUER` env var (default: `wayfarer`) |

## Token Verification

### Signing Algorithm

- **Algorithm**: HMAC-SHA256 (HS256)
- **Secret Key**: Value of `PLUGIN_LADDER_TO_HEAVEN_CRYPTEX_SECRET_KEY` environment variable

### Verification Steps

1. Decode the JWT token
2. Verify the signature using HMAC-SHA256 with the shared secret key
3. Check that `exp` (expiration) is in the future
4. Verify `aud` contains `"LADD-Cryptex"`
5. Optionally verify `iss` matches expected issuer

### Example Verification (Go)

```go
import (
    "github.com/golang-jwt/jwt/v5"
)

type CryptexClaims struct {
    UserName   string `json:"userName"`
    UserID     string `json:"userId"`
    ChurchID   string `json:"churchId"`
    ChurchName string `json:"churchName"`
    ProjectID  string `json:"projectId"`
    jwt.RegisteredClaims
}

func VerifyToken(tokenString, secretKey string) (*CryptexClaims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &CryptexClaims{}, func(token *jwt.Token) (interface{}, error) {
        // Verify signing method
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return []byte(secretKey), nil
    })

    if err != nil {
        return nil, err
    }

    claims, ok := token.Claims.(*CryptexClaims)
    if !ok || !token.Valid {
        return nil, fmt.Errorf("invalid token")
    }

    // Verify audience
    if !claims.VerifyAudience("LADD-Cryptex", true) {
        return nil, fmt.Errorf("invalid audience")
    }

    return claims, nil
}
```

### Example Verification (Node.js)

```javascript
const jwt = require('jsonwebtoken');

function verifyToken(tokenString, secretKey) {
    const decoded = jwt.verify(tokenString, secretKey, {
        algorithms: ['HS256'],
        audience: 'LADD-Cryptex'
    });

    return {
        userName: decoded.userName,
        userId: decoded.userId,
        churchId: decoded.churchId,
        churchName: decoded.churchName,
        projectId: decoded.projectId
    };
}
```

## Token Validity

- **Duration**: 6 months from issuance
- **Refresh**: Client should request a new token before expiration

## Configuration

The endpoint requires the following environment variables to be set on the Wayfarer backend:

```
PLUGIN_LADDER_TO_HEAVEN_CRYPTEX_SECRET_KEY=<shared-secret-key>
PLUGIN_LADDER_TO_HEAVEN_CRYPTEX_BASE_URL=<cryptex-base-url>
```

Example:
```
PLUGIN_LADDER_TO_HEAVEN_CRYPTEX_SECRET_KEY=my-secret-key-for-cryptex
PLUGIN_LADDER_TO_HEAVEN_CRYPTEX_BASE_URL=https://cryptex.example.com
```

The secret key must be configured in Cryptex for token verification.

## ID Format

Wayfarer uses prefixed ULIDs for all entity IDs:
- Format: `XX` + 26-character ULID = 28 characters total
- `US` prefix = User ID
- `CH` prefix = Church ID
- `PR` prefix = Project ID

Example: `US01ARZ3NDEKTSV4RRFFQ69G5FAV`

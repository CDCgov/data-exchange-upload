# Mock Authentication Server

## Overview

This is a lightweight **Mock Authentication Server** designed for testing OpenID Connect (OIDC) and JWT-based authentication flows. 

It provides:

- A mock **OpenID Provider** (`/.well-known/openid-configuration`)
- A **JWKS endpoint** (`/jwks`) for public key verification
- A **Token endpoint** (`/token`) that issues signed JWTs
- Multiple **Bad token endpoints** that issue tokens for testing edge cases
- Support for **configurable base URLs** via environment variables

This mock server was designed to be used by playwright and the upload-server to test authentication flows.

## Installation & Setup

### **1. Install Dependencies**

```sh
npm install
```

### **2. Configure Environment Variables**

Create a `.env` (or copy from `.env.example`) file in the root directory and define your settings:

```
MOCK_AUTH_BASE_URL=http://localhost
MOCK_AUTH_PORT=3000
```

- `MOCK_AUTH_BASE_URL` → The mock issuer URL (useful for testing different environments)
- `MOCK_AUTH_PORT` → The port to run the server on

### **3. Start the Server**

```sh
node server.js
```

## Usage

### **OpenID Configuration Endpoint**

Retrieve OpenID Connect metadata:

```sh
curl http://localhost:3000/.well-known/openid-configuration
```

### **JWKS (JSON Web Key Set) Endpoint**

Fetch the public key for verifying JWTs:

```sh
curl http://localhost:3000/oauth2/jwks
```

### **Token Endpoint (Generate JWTs)**

Request a valid access token:

```sh
curl http://localhost:3000/token
```

### **Bad Token Endpoint - Expired Token**

Request an access token that has an expired `exp` property set inside the token:

```sh
curl http://localhost:3000/token-expired
```

### **Bad Token Endpoint - Alternate Scopes Token**

Request an access token that provides alternative scopes for valdiating against unexpected scopes. Returns `test:scope3` and `test:scope4` as part of the `scope` in the JWT.

```sh
curl http://localhost:3000/token-scopes
```

## Environment Variables

| Variable             | Description                                      | Default                |
|----------------------|--------------------------------------------------|------------------------|
| `MOCK_AUTH_BASE_URL` | Base URL of the mock OpenID provider             | `http://localhost:3000`|
| `MOCK_AUTH_PORT`     | Port to run the mock server                      | `3000`                 |

## API Endpoints

| Method | Endpoint                                   | Description 
|--------|--------------------------------------------|------------
| GET    | `/.well-known/openid-configuration`        | Fetch OpenID Provider metadata
| GET    | `/jwks`                                    | Retrieve JWKS public keys
| GET    | `/token`                                   | Generate an access token
| GET    | `/token-expired`                           | Generate an access token that is expired
| GET    | `/token-scopes`                            | Generate an access token with alternate scopes

## Testing with Playwright

If you're using Playwright to test authentication, you can retrieve the token and use it for requests:

```js
        let token
        const res = await request.get('http://localhost:3000/token')
        const json = await res.json()
        token = json.access_token
```
# Database ERD - REST API

## Entity Relationship Diagram

```mermaid
erDiagram
    USERS ||--o{ REFRESH_TOKENS : "1:N"
    USERS ||--o{ NOTIFICATIONS : "1:N"
    USERS ||--o{ PUSH_TOKENS : "1:N"
    USERS ||--o{ PASSWORD_RESET_OTP : "1:N"
    USERS ||--o{ AUDIT_LOGS : "1:N"

    USERS {
        uint id PK
        string name
        string email UK
        string password
        timestamp expires_at
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
    }

    REFRESH_TOKENS {
        uint id PK
        uint user_id FK
        string token UK
        string token_id UK
        timestamp expires_at
        timestamp created_at
    }

    NOTIFICATIONS {
        uint id PK
        uint user_id FK "nullable"
        bool is_global
        string title
        string body
        string type "index"
        string data "json"
        bool is_read
        timestamp created_at
    }

    PUSH_TOKENS {
        uint id PK
        uint user_id FK
        string token UK
        string device
        timestamp created_at
    }

    PASSWORD_RESET_OTP {
        uint id PK
        uint user_id FK
        string otp UK
        timestamp expires_at
        timestamp created_at
    }

    TOKEN_BLACKLIST {
        uint id PK
        string token UK
        timestamp expires_at
        timestamp created_at
    }

    AUDIT_LOGS {
        uint id PK
        uint user_id FK
        string action
        string entity
        uint entity_id
        string details "json"
        string ip_address
        timestamp created_at
    }
```

## Database Schema Overview

```mermaid
graph TB
    subgraph users["users"]
        direction TB
        U1[id]
        U2[name]
        U3[email]
        U4[password]
        U5[expires_at]
        U6[created_at]
    end

    subgraph notifications["notifications"]
        direction TB
        N1[id]
        N2[user_id FK]
        N3[is_global]
        N4[title]
        N5[body]
        N6[type]
        N7[data json]
        N8[is_read]
        N9[created_at]
    end

    subgraph push_tokens["push_tokens"]
        direction TB
        P1[id]
        P2[user_id FK]
        P3[token]
        P4[device]
        P5[created_at]
    end

    subgraph refresh_tokens["refresh_tokens"]
        direction TB
        R1[id]
        R2[user_id FK]
        R3[token]
        R4[token_id]
        R5[expires_at]
        R6[created_at]
    end
```

## Table Relationships Flow

```mermaid
flowchart TD
    subgraph Core["Core Tables"]
        Users[users]
    end

    subgraph Auth["Authentication"]
        RefreshTokens[refresh_tokens]
        TokenBlacklist[token_blacklist]
        PasswordReset[password_reset_otp]
    end

    subgraph Notifications["Notifications"]
        Notifications[notifications]
        PushTokens[push_tokens]
    end

    subgraph Logging["Logging & Audit"]
        AuditLogs[audit_logs]
    end

    Users -->|1:N| RefreshTokens
    Users -->|1:N| TokenBlacklist
    Users -->|1:N| PasswordReset
    Users -->|1:N| Notifications
    Users -->|1:N| PushTokens
    Users -->|1:N| AuditLogs
```

## Foreign Key Reference Table

| Table | Foreign Key | References | Relationship |
|-------|-------------|------------|--------------|
| `refresh_tokens` | `user_id` | `users.id` | Many-to-One |
| `notifications` | `user_id` | `users.id` | Many-to-One |
| `push_tokens` | `user_id` | `users.id` | Many-to-One |
| `password_reset_otp` | `user_id` | `users.id` | Many-to-One |
| `audit_logs` | `user_id` | `users.id` | Many-to-One |

## Indexes Summary

| Table | Index | Type |
|-------|-------|------|
| `users` | email | Unique |
| `refresh_tokens` | user_id | Index |
| `refresh_tokens` | token | Unique |
| `refresh_tokens` | token_id | Unique |
| `notifications` | user_id | Index |
| `notifications` | type | Index |
| `push_tokens` | user_id | Index |
| `push_tokens` | token | Unique |
| `token_blacklist` | token | Unique |
| `password_reset_otp` | user_id | Index |
| `password_reset_otp` | otp | Unique |
| `audit_logs` | user_id | Index |

## Data Types Summary

```mermaid
flowchart LR
    subgraph StringTypes["String Types"]
        S1[name]
        S2[email]
        S3[password]
        S4[title]
        S5[body]
        S6[type]
        S7[data JSON]
        S8[token]
        S9[token_id]
        S10[otp]
        S11[action]
        S12[entity]
        S13[details JSON]
        S14[ip_address]
        S15[device]
    end

    subgraph NumericTypes["Numeric Types"]
        N1[uint]
    end

    subgraph DateTimeTypes["DateTime Types"]
        D1[expires_at]
        D2[created_at]
        D3[updated_at]
        D4[deleted_at]
    end

    subgraph BooleanTypes["Boolean Types"]
        B1[is_global]
        B2[is_read]
    end
```
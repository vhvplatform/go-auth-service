# Database Design Documentation

## Overview

The `go-auth-service` uses MongoDB as its primary database for storing authentication and authorization data. The database is designed to support multi-tenant SaaS applications with secure user authentication, role-based access control (RBAC), OAuth2 integration, and session management.

## Design Principles

1. **Multi-tenancy**: All tenant-specific data includes a `tenant_id` field for isolation
2. **Security**: Sensitive data (passwords) are hashed; tokens are securely stored
3. **Performance**: Strategic indexes for common query patterns
4. **Scalability**: Document-based design allows for flexible schema evolution
5. **Audit Trail**: Timestamps (created_at, updated_at) on all entities
6. **Data Integrity**: Unique constraints on critical fields

## Database Collections

### 1. users_auth

**Purpose**: Stores user authentication credentials and profile information.

**Schema**:
```json
{
  "_id": ObjectId,
  "email": String,
  "password_hash": String,
  "tenant_id": String,
  "roles": [String],
  "is_active": Boolean,
  "is_verified": Boolean,
  "last_login_at": Date (optional),
  "created_at": Date,
  "updated_at": Date
}
```

**Field Descriptions**:
- `_id`: Unique identifier (MongoDB ObjectID) - Primary Key
- `email`: User's email address - used for login
- `password_hash`: Bcrypt-hashed password (cost factor: 12)
- `tenant_id`: Reference to the tenant this user belongs to
- `roles`: Array of role names assigned to the user
- `is_active`: Flag indicating if account is active
- `is_verified`: Flag indicating if email is verified
- `last_login_at`: Timestamp of last successful login (nullable)
- `created_at`: Account creation timestamp
- `updated_at`: Last modification timestamp

**Indexes**:
1. `{email: 1}` - UNIQUE - Fast user lookup by email
2. `{tenant_id: 1}` - Filter users by tenant
3. `{email: 1, tenant_id: 1}` - UNIQUE - Ensures email uniqueness per tenant

**Constraints**:
- Email must be unique within a tenant
- Password hash must never be exposed in API responses (JSON tag: `-`)

**Design Decisions**:
- Email + tenant_id composite unique index allows same email across different tenants
- Roles stored as array of strings (denormalized) for fast authorization checks
- Password stored as bcrypt hash with appropriate cost factor for security
- is_active and is_verified flags enable account lifecycle management

---

### 2. refresh_tokens

**Purpose**: Stores refresh tokens for JWT authentication with revocation capability.

**Schema**:
```json
{
  "_id": ObjectId,
  "user_id": String,
  "token": String,
  "expires_at": Date,
  "created_at": Date,
  "revoked_at": Date (optional)
}
```

**Field Descriptions**:
- `_id`: Unique identifier (MongoDB ObjectID) - Primary Key
- `user_id`: Reference to user (users_auth._id as string)
- `token`: The actual refresh token string
- `expires_at`: Expiration timestamp (7 days default)
- `created_at`: Token creation timestamp
- `revoked_at`: Timestamp when token was revoked (null if active)

**Indexes**:
1. `{user_id: 1}` - Find all tokens for a user
2. `{token: 1}` - UNIQUE - Fast token validation
3. `{expires_at: 1}` - TTL index for automatic deletion

**Constraints**:
- Token must be unique across all refresh tokens
- TTL index automatically removes expired tokens

**Design Decisions**:
- user_id stored as string for flexibility (converts from ObjectID)
- revoked_at allows soft deletion and audit trail
- TTL index on expires_at enables automatic cleanup
- Token validation checks both expiry and revocation status

---

### 3. roles

**Purpose**: Defines roles and their associated permissions for RBAC.

**Schema**:
```json
{
  "_id": ObjectId,
  "name": String,
  "description": String,
  "permissions": [String],
  "tenant_id": String (optional),
  "created_at": Date,
  "updated_at": Date
}
```

**Field Descriptions**:
- `_id`: Unique identifier (MongoDB ObjectID) - Primary Key
- `name`: Role name (e.g., "admin", "user", "moderator")
- `description`: Human-readable description of the role
- `permissions`: Array of permission strings (e.g., "read:users", "write:posts")
- `tenant_id`: Tenant this role belongs to (null for system roles)
- `created_at`: Role creation timestamp
- `updated_at`: Last modification timestamp

**Indexes**:
1. `{name: 1, tenant_id: 1}` - UNIQUE - Ensures role names are unique per tenant

**Constraints**:
- Role name must be unique within a tenant
- System-wide roles have no tenant_id (or empty/null value)

**Design Decisions**:
- tenant_id is optional to support both system-wide and tenant-specific roles
- Permissions stored as array of strings for flexibility
- Composite unique index on (name, tenant_id) allows same role name across tenants
- Permission format follows "action:resource" pattern

---

### 4. permissions

**Purpose**: Defines available permissions in the system (optional collection for permission management).

**Schema**:
```json
{
  "_id": ObjectId,
  "name": String,
  "description": String,
  "resource": String,
  "action": String,
  "created_at": Date,
  "updated_at": Date
}
```

**Field Descriptions**:
- `_id`: Unique identifier (MongoDB ObjectID) - Primary Key
- `name`: Permission name (e.g., "read:users", "write:posts")
- `description`: Human-readable description
- `resource`: Resource type (e.g., "users", "posts", "settings")
- `action`: Action type (e.g., "read", "write", "delete", "admin")
- `created_at`: Permission creation timestamp
- `updated_at`: Last modification timestamp

**Indexes**:
- No specific indexes required (small collection, used for admin purposes)

**Design Decisions**:
- This collection is used for permission discovery and management UI
- Permissions are ultimately enforced through the roles.permissions array
- Granular separation of resource and action for flexible permission management

---

### 5. oauth_accounts

**Purpose**: Links OAuth provider accounts to local user accounts.

**Schema**:
```json
{
  "_id": ObjectId,
  "user_id": String,
  "provider": String,
  "provider_id": String,
  "email": String,
  "created_at": Date,
  "updated_at": Date
}
```

**Field Descriptions**:
- `_id`: Unique identifier (MongoDB ObjectID) - Primary Key
- `user_id`: Reference to user (users_auth._id as string)
- `provider`: OAuth provider name (e.g., "google", "github")
- `provider_id`: User's ID from the OAuth provider
- `email`: Email from OAuth provider
- `created_at`: Account linking timestamp
- `updated_at`: Last modification timestamp

**Indexes**:
- Primary index on _id (automatic)
- Recommended: `{user_id: 1}` - Find OAuth accounts for a user
- Recommended: `{provider: 1, provider_id: 1}` - UNIQUE - Ensure one link per provider account

**Design Decisions**:
- Allows one user to link multiple OAuth providers
- Email stored from provider for verification and matching
- Provider stored as enum-style string for type safety
- Supports account linking (connecting OAuth to existing account)

---

## Entity Relationships

### User ↔ RefreshToken (1:N)
- One user can have multiple refresh tokens (multiple sessions/devices)
- When user logs in, a new refresh token is created
- When user logs out, refresh token is revoked
- All refresh tokens are revoked on password change

### User ↔ Role (N:N)
- Users have an array of role names stored directly
- Roles are looked up from the roles collection
- Many users can have the same role
- One user can have multiple roles

### Role ↔ Permission (1:N embedded)
- Roles contain an array of permission strings
- Permissions are embedded in roles for performance
- No direct foreign key relationship (denormalized)

### User ↔ OAuthAccount (1:N)
- One user can link multiple OAuth providers
- Each OAuth account links to exactly one user
- OAuth accounts can be created before or after user registration

### Tenant Relationships
- Users belong to one tenant (tenant_id field)
- Roles can be tenant-specific or system-wide
- Isolation enforced at application layer through queries

---

## Indexes and Query Optimization

### Index Strategy

1. **users_auth**:
   - Unique composite index on (email, tenant_id) for login queries
   - Single field index on email for global lookup
   - Index on tenant_id for tenant-wide user queries

2. **refresh_tokens**:
   - Unique index on token for fast validation
   - Index on user_id for user token management
   - TTL index on expires_at for automatic cleanup

3. **roles**:
   - Unique composite index on (name, tenant_id) for role lookup

### Query Patterns

**Most Frequent Queries**:
1. User login: `{email: "...", tenant_id: "..."}`
2. Token validation: `{token: "...", revoked_at: null, expires_at: {$gt: now}}`
3. Role permission lookup: `{name: {$in: [...roles]}, tenant_id: "..."}`

**Performance Considerations**:
- All primary lookup queries are covered by indexes
- Composite indexes support both exact match and prefix queries
- TTL indexes reduce manual cleanup operations
- Denormalized role array in users speeds up authorization

---

## Security Considerations

### Password Security
- Passwords hashed with bcrypt (cost factor: 12)
- Password hashes never returned in API responses
- Password validation timing is constant to prevent timing attacks
- Password requirements enforced at application layer

### Token Security
- Refresh tokens are unique and unpredictable
- Tokens are revocable (revoked_at timestamp)
- Expired tokens automatically deleted by TTL index
- Token blacklist maintained in Redis for access tokens

### Multi-Tenancy
- All queries include tenant_id filter for isolation
- Cross-tenant data access prevented at application layer
- Unique constraints scoped per tenant where applicable

### Audit Trail
- All entities include created_at and updated_at timestamps
- Token revocation tracked with revoked_at
- Last login timestamp tracked for security monitoring

---

## Data Lifecycle

### User Lifecycle
1. **Registration**: User created with is_verified=false
2. **Verification**: is_verified set to true after email verification
3. **Activation**: is_active controls account state
4. **Login**: last_login_at updated on successful authentication
5. **Password Change**: All refresh tokens revoked
6. **Deactivation**: is_active set to false (soft delete)

### Token Lifecycle
1. **Creation**: Refresh token created on login/refresh
2. **Validation**: Checked against expiry and revocation
3. **Refresh**: Old token revoked, new token created
4. **Revocation**: revoked_at set on logout
5. **Expiration**: TTL index automatically removes expired tokens

### Role Lifecycle
1. **Creation**: System or tenant admin creates role
2. **Assignment**: Role names added to user.roles array
3. **Update**: Permissions modified on role
4. **Usage**: Permissions loaded on authorization check
5. **Deletion**: Role removed, references remain in user.roles

---

## Scalability and Performance

### Current Design Strengths
- Document-based storage allows flexible schema evolution
- Embedded permissions in roles reduce joins
- Strategic indexes optimize common queries
- TTL indexes automate cleanup tasks

### Scaling Strategies
1. **Horizontal Scaling**: MongoDB replica sets for read scaling
2. **Sharding**: Shard by tenant_id for multi-tenant scaling
3. **Caching**: Redis cache for user sessions and frequently accessed data
4. **Index Optimization**: Monitor and add indexes based on query patterns

### Monitoring Recommendations
- Track slow queries (>100ms)
- Monitor index usage with MongoDB's explain()
- Set up alerts for collection size growth
- Monitor TTL index effectiveness

---

## Migration Considerations

### Adding New Fields
- MongoDB's flexible schema allows adding fields without migrations
- Use default values in application code for backward compatibility
- Consider gradual rollout for mandatory fields

### Changing Indexes
1. Create new index online (non-blocking)
2. Deploy application code using new index
3. Remove old index after verification

### Data Cleanup
- Use TTL indexes for automatic cleanup
- Implement background jobs for complex cleanup logic
- Consider archival strategy for audit data

---

## Future Enhancements

### Potential Improvements
1. **Audit Logs**: Dedicated collection for security events
2. **Password History**: Prevent password reuse
3. **Account Linking**: Enhanced OAuth account management
4. **Session Management**: Dedicated sessions collection
5. **Device Tracking**: Store device info with refresh tokens
6. **Permission Hierarchies**: Support for permission inheritance

### Monitoring and Analytics
1. Login patterns and trends
2. Failed authentication attempts
3. Token usage statistics
4. Role and permission usage analytics

---

## References

- MongoDB Indexing Best Practices: https://docs.mongodb.com/manual/indexes/
- MongoDB Schema Design Patterns: https://www.mongodb.com/blog/post/building-with-patterns-a-summary
- OWASP Authentication Cheat Sheet: https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html
- JWT Best Practices: https://tools.ietf.org/html/rfc8725

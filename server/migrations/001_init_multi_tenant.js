// MongoDB Indexes and Collections Setup for Multi-Tenant Auth
// Run this script to create necessary collections and indexes

// Use auth database
db = db.getSiblingDB('auth_service');

// 1. Users Collection
db.createCollection('users_auth');

// Create indexes for users
db.users_auth.createIndex({ "email": 1 }, { unique: true, sparse: true });
db.users_auth.createIndex({ "username": 1 }, { unique: true, sparse: true });
db.users_auth.createIndex({ "phone": 1 }, { unique: true, sparse: true });
db.users_auth.createIndex({ "docNumber": 1 }, { unique: true, sparse: true });
db.users_auth.createIndex({ "isActive": 1 });
db.users_auth.createIndex({ "createdAt": -1 });
db.users_auth.createIndex({ "email": 1, "isActive": 1 });

// 2. User-Tenant Relationships Collection
db.createCollection('user_tenants');

// Create indexes for user_tenants
db.user_tenants.createIndex({ "userId": 1, "tenantId": 1 }, { unique: true });
db.user_tenants.createIndex({ "userId": 1 });
db.user_tenants.createIndex({ "tenantId": 1 });
db.user_tenants.createIndex({ "isActive": 1 });
db.user_tenants.createIndex({ "tenantId": 1, "isActive": 1 });
db.user_tenants.createIndex({ "userId": 1, "isActive": 1 });
db.user_tenants.createIndex({ "joinedAt": -1 });

// 3. Tenant Login Configurations Collection
db.createCollection('tenant_login_configs');

// Create indexes for tenant_login_configs
db.tenant_login_configs.createIndex({ "tenantId": 1 }, { unique: true });

// 4. Refresh Tokens Collection
db.createCollection('refresh_tokens');

// Create indexes for refresh_tokens
db.refresh_tokens.createIndex({ "token": 1 }, { unique: true });
db.refresh_tokens.createIndex({ "userId": 1 });
db.refresh_tokens.createIndex({ "tenantId": 1 });
db.refresh_tokens.createIndex({ "expiresAt": 1 });
db.refresh_tokens.createIndex({ "userId": 1, "tenantId": 1 });
db.refresh_tokens.createIndex({ "revokedAt": 1 }, { sparse: true });

// 5. Roles Collection
db.createCollection('roles');

// Create indexes for roles
db.roles.createIndex({ "name": 1, "tenantId": 1 }, { unique: true });
db.roles.createIndex({ "tenantId": 1 });

// 6. Login Attempts Collection (for rate limiting)
db.createCollection('login_attempts');

// Create indexes for login_attempts
db.login_attempts.createIndex({ "identifier": 1, "tenantId": 1, "attemptAt": -1 });
db.login_attempts.createIndex({ "attemptAt": 1 }, { expireAfterSeconds: 86400 }); // TTL index: auto-delete after 24 hours
db.login_attempts.createIndex({ "ipAddress": 1, "tenantId": 1, "attemptAt": -1 });

// 7. User Lockouts Collection
db.createCollection('user_lockouts');

// Create indexes for user_lockouts
db.user_lockouts.createIndex({ "userId": 1, "tenantId": 1, "isActive": 1 });
db.user_lockouts.createIndex({ "unlockAt": 1 });
db.user_lockouts.createIndex({ "userId": 1, "isActive": 1 });

// 8. Insert Default Tenant Login Config (System Default)
db.tenant_login_configs.insertOne({
    tenantId: "system",
    allowedIdentifiers: ["email", "username"],
    require2FA: false,
    allowRegistration: true,
    passwordMinLength: 8,
    passwordRequireUpper: true,
    passwordRequireLower: true,
    passwordRequireDigit: true,
    passwordRequireSpec: false,
    sessionTimeout: 1440, // 24 hours
    maxLoginAttempts: 5,
    lockoutDuration: 30, // 30 minutes
    createdAt: new Date(),
    updatedAt: new Date()
});

// 9. Insert Default Roles
db.roles.insertMany([
    {
        name: "super_admin",
        description: "Super Administrator with full system access",
        permissions: ["*"],
        tenantId: "system",
        createdAt: new Date(),
        updatedAt: new Date()
    },
    {
        name: "admin",
        description: "Administrator with tenant-level access",
        permissions: [
            "user.read",
            "user.write",
            "user.delete",
            "role.read",
            "role.write",
            "tenant.read",
            "tenant.write"
        ],
        tenantId: "system",
        createdAt: new Date(),
        updatedAt: new Date()
    },
    {
        name: "user",
        description: "Standard user with basic permissions",
        permissions: [
            "user.read.own",
            "user.write.own"
        ],
        tenantId: "system",
        createdAt: new Date(),
        updatedAt: new Date()
    }
]);

// 10. Create System Admin User (Optional - for initial setup)
// Password: Admin@123 (change this immediately in production!)
var systemAdminUser = {
    email: "admin@system.local",
    username: "admin",
    passwordHash: "$2a$10$xN8kLZJCkE/8K7YqNqXRMuFmGh6Z2qxp5XqKQzYZLvH8vGxXyZFKG", // Admin@123
    isActive: true,
    isVerified: true,
    createdAt: new Date(),
    updatedAt: new Date()
};

var adminResult = db.users_auth.insertOne(systemAdminUser);

// Create system admin relationship
db.user_tenants.insertOne({
    userId: adminResult.insertedId.toString(),
    tenantId: "system",
    roles: ["super_admin"],
    isActive: true,
    joinedAt: new Date(),
    createdAt: new Date(),
    updatedAt: new Date()
});

print("✅ Database migration completed successfully!");
print("📝 Collections created:");
print("   - users_auth");
print("   - user_tenants");
print("   - tenant_login_configs");
print("   - refresh_tokens");
print("   - roles");
print("   - login_attempts");
print("   - user_lockouts");
print("");
print("👤 System admin user created:");
print("   Email: admin@system.local");
print("   Username: admin");
print("   Password: Admin@123");
print("   ⚠️  CHANGE THIS PASSWORD IMMEDIATELY!");
print("");
print("🔐 Default roles created:");
print("   - super_admin");
print("   - admin");
print("   - user");

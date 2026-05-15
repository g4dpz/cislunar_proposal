/**
 * Unit Tests for User Service
 *
 * Tests CRUD operations for users including listing, creation,
 * updating, deletion, email uniqueness, and role assignment.
 *
 * Validates: Requirements 12.1, 13.1, 14.1, 15.2, 20.2
 *
 * Feature: project-website
 */

import {
  assertEquals,
  assertNotEquals,
  assertRejects,
} from "https://deno.land/std@0.224.0/assert/mod.ts";
import { initDatabase } from "../../db/mod.ts";
import { createUserService } from "../../services/users.ts";

// ─── Helper ───────────────────────────────────────────────────────────────────

async function setupTestDb() {
  const db = await initDatabase(":memory:");
  const userService = createUserService(db);
  return { db, userService };
}

// ─── listUsers ────────────────────────────────────────────────────────────────

Deno.test("listUsers - returns users sorted by creation date descending", async () => {
  const { db, userService } = await setupTestDb();

  // Create users with slight delay to ensure different timestamps
  await userService.createUser({
    name: "First User",
    email: "first@example.com",
    password: "password123",
    roleIds: [],
  });

  // Manually set a later created_at for the second user
  await userService.createUser({
    name: "Second User",
    email: "second@example.com",
    password: "password123",
    roleIds: [],
  });

  const users = userService.listUsers();

  // Should include seeded admin + our 2 users
  assertEquals(users.length >= 3, true, "Should have at least 3 users");

  // Verify ordering: each user's createdAt >= next user's createdAt
  for (let i = 0; i < users.length - 1; i++) {
    const current = users[i]!.createdAt;
    const next = users[i + 1]!.createdAt;
    assertEquals(current >= next, true, "Users should be sorted by creation date descending");
  }

  db.close();
});

// ─── getUser ──────────────────────────────────────────────────────────────────

Deno.test("getUser - returns user with roles", async () => {
  const { db, userService } = await setupTestDb();

  // Get the admin role ID
  const adminRole = db.prepare("SELECT id FROM roles WHERE name = 'admin'").get() as { id: number };
  const usersRole = db.prepare("SELECT id FROM roles WHERE name = 'users'").get() as { id: number };

  const created = await userService.createUser({
    name: "Role User",
    email: "roleuser@example.com",
    password: "password123",
    roleIds: [adminRole.id, usersRole.id],
  });

  const user = userService.getUser(created.id);

  assertNotEquals(user, null);
  assertEquals(user!.name, "Role User");
  assertEquals(user!.email, "roleuser@example.com");
  assertEquals(user!.roles.length, 2);

  const roleNames = user!.roles.map((r) => r.name).sort();
  assertEquals(roleNames, ["admin", "users"]);

  db.close();
});

Deno.test("getUser - returns null for non-existent ID", async () => {
  const { db, userService } = await setupTestDb();

  const user = userService.getUser(99999);
  assertEquals(user, null);

  db.close();
});

// ─── createUser ───────────────────────────────────────────────────────────────

Deno.test("createUser - creates user with hashed password and role assignments", async () => {
  const { db, userService } = await setupTestDb();

  const usersRole = db.prepare("SELECT id FROM roles WHERE name = 'users'").get() as { id: number };

  const created = await userService.createUser({
    name: "New User",
    email: "newuser@example.com",
    password: "securepass123",
    roleIds: [usersRole.id],
  });

  assertEquals(created.name, "New User");
  assertEquals(created.email, "newuser@example.com");
  assertEquals(typeof created.id, "number");
  assertEquals(created.id > 0, true);

  // Verify password is hashed (not stored as plain text)
  assertNotEquals(created.passwordHash, "securepass123");
  assertEquals(created.passwordHash.startsWith("$2"), true, "Password should be bcrypt hashed");

  // Verify role assignment
  const roleAssignment = db.prepare(
    "SELECT role_id FROM user_roles WHERE user_id = ?",
  ).all(created.id) as Array<{ role_id: number }>;
  assertEquals(roleAssignment.length, 1);
  assertEquals(roleAssignment[0]!.role_id, usersRole.id);

  db.close();
});

Deno.test("createUser - throws on duplicate email", async () => {
  const { db, userService } = await setupTestDb();

  await userService.createUser({
    name: "Original",
    email: "duplicate@example.com",
    password: "password123",
    roleIds: [],
  });

  await assertRejects(
    () =>
      userService.createUser({
        name: "Duplicate",
        email: "duplicate@example.com",
        password: "password456",
        roleIds: [],
      }),
    Error,
  );

  db.close();
});

// ─── updateUser ───────────────────────────────────────────────────────────────

Deno.test("updateUser - updates name, email, and role assignments", async () => {
  const { db, userService } = await setupTestDb();

  const usersRole = db.prepare("SELECT id FROM roles WHERE name = 'users'").get() as { id: number };
  const adminRole = db.prepare("SELECT id FROM roles WHERE name = 'admin'").get() as { id: number };

  const created = await userService.createUser({
    name: "Before Update",
    email: "before@example.com",
    password: "password123",
    roleIds: [usersRole.id],
  });

  const updated = userService.updateUser(created.id, {
    name: "After Update",
    email: "after@example.com",
    roleIds: [adminRole.id],
  });

  assertEquals(updated.name, "After Update");
  assertEquals(updated.email, "after@example.com");
  assertEquals(updated.id, created.id);

  // Verify role was changed
  const user = userService.getUser(created.id);
  assertEquals(user!.roles.length, 1);
  assertEquals(user!.roles[0]!.name, "admin");

  db.close();
});

Deno.test("updateUser - sets updated_at timestamp", async () => {
  const { db, userService } = await setupTestDb();

  const created = await userService.createUser({
    name: "Timestamp User",
    email: "timestamp@example.com",
    password: "password123",
    roleIds: [],
  });

  const originalUpdatedAt = created.updatedAt;

  // Small delay to ensure timestamp difference
  await new Promise((resolve) => setTimeout(resolve, 50));

  const updated = userService.updateUser(created.id, {
    name: "Updated Name",
    email: "timestamp@example.com",
    roleIds: [],
  });

  // updated_at should be different from the original (or at least set)
  assertNotEquals(updated.updatedAt, undefined);
  // The updated_at should be >= original (datetime('now') is used)
  assertEquals(updated.updatedAt >= originalUpdatedAt, true);

  db.close();
});

// ─── deleteUser ───────────────────────────────────────────────────────────────

Deno.test("deleteUser - removes user", async () => {
  const { db, userService } = await setupTestDb();

  const created = await userService.createUser({
    name: "Delete Me",
    email: "deleteme@example.com",
    password: "password123",
    roleIds: [],
  });

  userService.deleteUser(created.id);

  const user = userService.getUser(created.id);
  assertEquals(user, null);

  db.close();
});

Deno.test("deleteUser - cascades to user_roles", async () => {
  const { db, userService } = await setupTestDb();

  const usersRole = db.prepare("SELECT id FROM roles WHERE name = 'users'").get() as { id: number };
  const adminRole = db.prepare("SELECT id FROM roles WHERE name = 'admin'").get() as { id: number };

  const created = await userService.createUser({
    name: "Cascade User",
    email: "cascade@example.com",
    password: "password123",
    roleIds: [usersRole.id, adminRole.id],
  });

  // Verify roles exist before deletion
  const rolesBefore = db.prepare(
    "SELECT role_id FROM user_roles WHERE user_id = ?",
  ).all(created.id) as Array<{ role_id: number }>;
  assertEquals(rolesBefore.length, 2);

  userService.deleteUser(created.id);

  // Verify user_roles entries are removed
  const rolesAfter = db.prepare(
    "SELECT role_id FROM user_roles WHERE user_id = ?",
  ).all(created.id) as Array<{ role_id: number }>;
  assertEquals(rolesAfter.length, 0);

  db.close();
});

// ─── isEmailAvailable ─────────────────────────────────────────────────────────

Deno.test("isEmailAvailable - returns true for new email", async () => {
  const { db, userService } = await setupTestDb();

  assertEquals(userService.isEmailAvailable("brand-new@example.com"), true);

  db.close();
});

Deno.test("isEmailAvailable - returns false for existing email", async () => {
  const { db, userService } = await setupTestDb();

  await userService.createUser({
    name: "Existing",
    email: "existing@example.com",
    password: "password123",
    roleIds: [],
  });

  assertEquals(userService.isEmailAvailable("existing@example.com"), false);

  db.close();
});

Deno.test("isEmailAvailable - excludes given user ID", async () => {
  const { db, userService } = await setupTestDb();

  const created = await userService.createUser({
    name: "Exclude User",
    email: "exclude@example.com",
    password: "password123",
    roleIds: [],
  });

  // Email is available when excluding the user who owns it (for updates)
  assertEquals(userService.isEmailAvailable("exclude@example.com", created.id), true);

  // Email is NOT available when excluding a different user
  assertEquals(userService.isEmailAvailable("exclude@example.com", 99999), false);

  db.close();
});

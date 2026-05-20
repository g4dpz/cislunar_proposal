/**
 * Unit Tests for Role Service
 *
 * Tests CRUD operations for roles including listing, creation,
 * updating, deletion constraints, and name availability checks.
 *
 * Feature: project-website
 */

import {
  assertEquals,
  assertNotEquals,
} from "https://deno.land/std@0.224.0/assert/mod.ts";
import { initDatabase } from "../../db/mod.ts";
import { createRoleService } from "../../services/roles.ts";

// ─── listRoles ────────────────────────────────────────────────────────────────

Deno.test("listRoles - returns seeded roles with user counts", async () => {
  const db = await initDatabase(":memory:");
  const roleService = createRoleService(db);

  const roles = roleService.listRoles();

  // Seeded roles: "admin" and "users"
  assertEquals(roles.length >= 2, true, "Should have at least 2 seeded roles");

  const adminRole = roles.find((r) => r.name === "admin");
  const usersRole = roles.find((r) => r.name === "users");

  assertNotEquals(adminRole, undefined, "admin role should exist");
  assertNotEquals(usersRole, undefined, "users role should exist");

  // Admin user is seeded and assigned the admin role
  assertEquals(adminRole!.userCount >= 1, true, "admin role should have at least 1 user");

  db.close();
});

Deno.test("listRoles - includes user count from user_roles", async () => {
  const db = await initDatabase(":memory:");
  const roleService = createRoleService(db);

  // Create a new role with no users
  const newRole = roleService.createRole({
    name: "testers",
    description: "Testing role",
  });

  const roles = roleService.listRoles();
  const testRole = roles.find((r) => r.id === newRole.id);

  assertNotEquals(testRole, undefined);
  assertEquals(testRole!.userCount, 0, "New role should have 0 users");

  db.close();
});

// ─── getRole ──────────────────────────────────────────────────────────────────

Deno.test("getRole - returns role with assigned users", async () => {
  const db = await initDatabase(":memory:");
  const roleService = createRoleService(db);

  // Get the seeded admin role
  const roles = roleService.listRoles();
  const adminRole = roles.find((r) => r.name === "admin")!;

  const role = roleService.getRole(adminRole.id);

  assertNotEquals(role, null, "Should find the admin role");
  assertEquals(role!.name, "admin");
  assertEquals(role!.users.length >= 1, true, "Admin role should have at least 1 user");
  assertEquals(role!.users[0]!.email, "admin@radiant.radio");

  db.close();
});

Deno.test("getRole - returns null for non-existent role", async () => {
  const db = await initDatabase(":memory:");
  const roleService = createRoleService(db);

  const role = roleService.getRole(9999);
  assertEquals(role, null);

  db.close();
});

// ─── createRole ───────────────────────────────────────────────────────────────

Deno.test("createRole - creates role with name and description", async () => {
  const db = await initDatabase(":memory:");
  const roleService = createRoleService(db);

  const role = roleService.createRole({
    name: "moderator",
    description: "Can moderate content",
  });

  assertEquals(role.name, "moderator");
  assertEquals(role.description, "Can moderate content");
  assertEquals(typeof role.id, "number");
  assertEquals(role.id > 0, true);

  db.close();
});

// ─── updateRole ───────────────────────────────────────────────────────────────

Deno.test("updateRole - updates name and description", async () => {
  const db = await initDatabase(":memory:");
  const roleService = createRoleService(db);

  const created = roleService.createRole({
    name: "editor",
    description: "Can edit content",
  });

  const updated = roleService.updateRole(created.id, {
    name: "senior-editor",
    description: "Can edit and publish content",
  });

  assertEquals(updated.id, created.id);
  assertEquals(updated.name, "senior-editor");
  assertEquals(updated.description, "Can edit and publish content");

  // Verify via getRole
  const fetched = roleService.getRole(created.id);
  assertEquals(fetched!.name, "senior-editor");
  assertEquals(fetched!.description, "Can edit and publish content");

  db.close();
});

// ─── deleteRole ───────────────────────────────────────────────────────────────

Deno.test("deleteRole - deletes role with no assigned users", async () => {
  const db = await initDatabase(":memory:");
  const roleService = createRoleService(db);

  const role = roleService.createRole({
    name: "temporary",
    description: "Temporary role",
  });

  const result = roleService.deleteRole(role.id);
  assertEquals(result.success, true);
  assertEquals(result.error, undefined);

  // Verify role is gone
  const fetched = roleService.getRole(role.id);
  assertEquals(fetched, null);

  db.close();
});

Deno.test("deleteRole - fails when users are assigned", async () => {
  const db = await initDatabase(":memory:");
  const roleService = createRoleService(db);

  // The admin role has the seeded admin user assigned
  const roles = roleService.listRoles();
  const adminRole = roles.find((r) => r.name === "admin")!;

  const result = roleService.deleteRole(adminRole.id);
  assertEquals(result.success, false);
  assertEquals(result.error, "Cannot delete role with assigned users");

  // Verify role still exists
  const fetched = roleService.getRole(adminRole.id);
  assertNotEquals(fetched, null);

  db.close();
});

// ─── isNameAvailable ──────────────────────────────────────────────────────────

Deno.test("isNameAvailable - returns false for existing name", async () => {
  const db = await initDatabase(":memory:");
  const roleService = createRoleService(db);

  assertEquals(roleService.isNameAvailable("admin"), false);
  assertEquals(roleService.isNameAvailable("users"), false);

  db.close();
});

Deno.test("isNameAvailable - returns true for new name", async () => {
  const db = await initDatabase(":memory:");
  const roleService = createRoleService(db);

  assertEquals(roleService.isNameAvailable("brand-new-role"), true);

  db.close();
});

Deno.test("isNameAvailable - excludes given role ID", async () => {
  const db = await initDatabase(":memory:");
  const roleService = createRoleService(db);

  // Get the admin role ID
  const roles = roleService.listRoles();
  const adminRole = roles.find((r) => r.name === "admin")!;

  // "admin" is available if we exclude the admin role itself (for updates)
  assertEquals(roleService.isNameAvailable("admin", adminRole.id), true);

  // "admin" is NOT available if we exclude a different role
  assertEquals(roleService.isNameAvailable("admin", 9999), false);

  db.close();
});

/**
 * Unit tests for services/auth.ts
 *
 * Tests registration, login, logout, session validation, password change,
 * and expired session cleanup.
 */

import {
  assertEquals,
  assertNotEquals,
  assertRejects,
} from "https://deno.land/std@0.224.0/assert/mod.ts";
import { initDatabase } from "../../db/mod.ts";
import { createAuthService } from "../../services/auth.ts";

async function setupTestDb() {
  const db = await initDatabase(":memory:");
  const authService = createAuthService(db);
  return { db, authService };
}

Deno.test("register - creates user with session and assigns 'users' role", async () => {
  const { db, authService } = await setupTestDb();

  const result = await authService.register("Test User", "test@example.com", "password123");

  assertNotEquals(result.sessionToken, "");
  assertEquals(result.user.name, "Test User");
  assertEquals(result.user.email, "test@example.com");

  // Verify session exists in database
  const session = db.prepare("SELECT id FROM sessions WHERE id = ?").get(result.sessionToken);
  assertNotEquals(session, undefined);

  // Verify "users" role is assigned
  const roleAssignment = db.prepare(
    `SELECT r.name FROM roles r
     INNER JOIN user_roles ur ON ur.role_id = r.id
     WHERE ur.user_id = ?`
  ).get(result.user.id) as { name: string } | undefined;
  assertEquals(roleAssignment?.name, "users");

  db.close();
});

Deno.test("register - rejects password shorter than 8 characters", async () => {
  const { db, authService } = await setupTestDb();

  await assertRejects(
    () => authService.register("Test", "short@example.com", "short"),
    Error,
    "Password must be at least 8 characters",
  );

  db.close();
});

Deno.test("register - rejects duplicate email", async () => {
  const { db, authService } = await setupTestDb();

  await authService.register("User 1", "dup@example.com", "password123");

  await assertRejects(
    () => authService.register("User 2", "dup@example.com", "password456"),
    Error,
    "Email is already registered",
  );

  db.close();
});

Deno.test("login - returns session token for valid credentials", async () => {
  const { db, authService } = await setupTestDb();

  await authService.register("Login User", "login@example.com", "mypassword1");
  const result = await authService.login("login@example.com", "mypassword1");

  assertNotEquals(result, null);
  assertNotEquals(result!.sessionToken, "");
  assertEquals(result!.user.email, "login@example.com");

  db.close();
});

Deno.test("login - returns null for wrong password", async () => {
  const { db, authService } = await setupTestDb();

  await authService.register("User", "wrong@example.com", "correctpass");
  const result = await authService.login("wrong@example.com", "wrongpass!!");

  assertEquals(result, null);

  db.close();
});

Deno.test("login - returns null for non-existent email", async () => {
  const { db, authService } = await setupTestDb();

  const result = await authService.login("nobody@example.com", "password123");
  assertEquals(result, null);

  db.close();
});

Deno.test("logout - removes session from database", async () => {
  const { db, authService } = await setupTestDb();

  const { sessionToken } = await authService.register("User", "logout@example.com", "password123");

  // Session exists before logout
  const before = db.prepare("SELECT id FROM sessions WHERE id = ?").get(sessionToken);
  assertNotEquals(before, undefined);

  await authService.logout(sessionToken);

  // Session removed after logout
  const after = db.prepare("SELECT id FROM sessions WHERE id = ?").get(sessionToken);
  assertEquals(after, undefined);

  db.close();
});

Deno.test("validateSession - returns user with roles for valid session", async () => {
  const { db, authService } = await setupTestDb();

  const { sessionToken } = await authService.register("Valid User", "valid@example.com", "password123");
  const userWithRoles = await authService.validateSession(sessionToken);

  assertNotEquals(userWithRoles, null);
  assertEquals(userWithRoles!.name, "Valid User");
  assertEquals(userWithRoles!.email, "valid@example.com");
  assertEquals(userWithRoles!.roles.length, 1);
  assertEquals(userWithRoles!.roles[0]!.name, "users");

  db.close();
});

Deno.test("validateSession - returns null for non-existent session", async () => {
  const { db, authService } = await setupTestDb();

  const result = await authService.validateSession("non-existent-token");
  assertEquals(result, null);

  db.close();
});

Deno.test("validateSession - returns null for expired session", async () => {
  const { db, authService } = await setupTestDb();

  const { sessionToken, user } = await authService.register("Expired User", "expired@example.com", "password123");

  // Manually set session to expired (1 hour in the past)
  const pastTime = new Date(Date.now() - 60 * 60 * 1000 - 1000).toISOString();
  db.exec("UPDATE sessions SET expires_at = ? WHERE id = ?", [pastTime, sessionToken]);

  const result = await authService.validateSession(sessionToken);
  assertEquals(result, null);

  db.close();
});

Deno.test("changePassword - updates password when current password is correct", async () => {
  const { db, authService } = await setupTestDb();

  const { user } = await authService.register("PW User", "pw@example.com", "oldpassword1");

  const changed = await authService.changePassword(user.id, "oldpassword1", "newpassword1");
  assertEquals(changed, true);

  // Verify new password works for login
  const loginResult = await authService.login("pw@example.com", "newpassword1");
  assertNotEquals(loginResult, null);

  // Verify old password no longer works
  const oldLogin = await authService.login("pw@example.com", "oldpassword1");
  assertEquals(oldLogin, null);

  db.close();
});

Deno.test("changePassword - returns false for incorrect current password", async () => {
  const { db, authService } = await setupTestDb();

  const { user } = await authService.register("PW User", "pw2@example.com", "mypassword1");

  const changed = await authService.changePassword(user.id, "wrongcurrent", "newpassword1");
  assertEquals(changed, false);

  db.close();
});

Deno.test("changePassword - rejects new password shorter than 8 characters", async () => {
  const { db, authService } = await setupTestDb();

  const { user } = await authService.register("PW User", "pw3@example.com", "mypassword1");

  await assertRejects(
    () => authService.changePassword(user.id, "mypassword1", "short"),
    Error,
    "Password must be at least 8 characters",
  );

  db.close();
});

Deno.test("cleanExpiredSessions - removes only expired sessions", async () => {
  const { db, authService } = await setupTestDb();

  // Create a user and two sessions
  const { user } = await authService.register("Clean User", "clean@example.com", "password123");

  // Insert an expired session
  const expiredToken = crypto.randomUUID();
  const pastTime = new Date(Date.now() - 60 * 60 * 1000 - 1000).toISOString();
  db.exec(
    "INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)",
    [expiredToken, user.id, pastTime],
  );

  // Insert a valid session
  const validToken = crypto.randomUUID();
  const futureTime = new Date(Date.now() + 60 * 60 * 1000).toISOString();
  db.exec(
    "INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)",
    [validToken, user.id, futureTime],
  );

  await authService.cleanExpiredSessions();

  // Expired session should be gone
  const expired = db.prepare("SELECT id FROM sessions WHERE id = ?").get(expiredToken);
  assertEquals(expired, undefined);

  // Valid session should still exist
  const valid = db.prepare("SELECT id FROM sessions WHERE id = ?").get(validToken);
  assertNotEquals(valid, undefined);

  db.close();
});

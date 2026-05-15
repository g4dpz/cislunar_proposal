/**
 * Property 7: Database Seeding Idempotence
 *
 * Running the database seed function N times (N ≥ 1) produces the same state
 * as running it once: exactly 2 default roles and 1 admin user exist.
 *
 * Feature: project-website
 * Property 7: Database Seeding Idempotence
 * Validates: Requirements 27.1, 27.4
 */

import { assertEquals } from "https://deno.land/std@0.224.0/assert/mod.ts";
import fc from "fast-check";
import { initDatabase, seedDatabase } from "../../db/mod.ts";

Deno.test("Property 7: Database Seeding Idempotence - seeding N times produces same state as seeding once", async () => {
  await fc.assert(
    fc.asyncProperty(fc.integer({ min: 1, max: 10 }), async (n: number) => {
      // Create a fresh in-memory database and run initDatabase (which calls seedDatabase once)
      const db = await initDatabase(":memory:");

      // Call seedDatabase N-1 more times (total of N seed calls)
      for (let i = 1; i < n; i++) {
        await seedDatabase(db);
      }

      // Verify exactly 2 roles exist ("admin" and "users")
      const roles = db.prepare("SELECT name FROM roles ORDER BY name").all() as Array<{ name: string }>;
      assertEquals(roles.length, 2, `Expected exactly 2 roles, got ${roles.length} after ${n} seed calls`);
      assertEquals(roles[0]!.name, "admin");
      assertEquals(roles[1]!.name, "users");

      // Verify exactly 1 user exists with email "admin@arthur.radio"
      const users = db.prepare("SELECT email FROM users").all() as Array<{ email: string }>;
      assertEquals(users.length, 1, `Expected exactly 1 user, got ${users.length} after ${n} seed calls`);
      assertEquals(users[0]!.email, "admin@arthur.radio");

      // Verify exactly 1 user_roles entry exists (admin user → admin role)
      const userRoles = db.prepare("SELECT user_id, role_id FROM user_roles").all() as Array<{ user_id: number; role_id: number }>;
      assertEquals(userRoles.length, 1, `Expected exactly 1 user_roles entry, got ${userRoles.length} after ${n} seed calls`);

      // Clean up
      db.close();
    }),
    { numRuns: 20 }
  );
});

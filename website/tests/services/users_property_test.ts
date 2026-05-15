/**
 * Property-Based Tests for User Service
 *
 * Property 1: Email Uniqueness Invariant
 * Property 4: User CRUD Round-Trip
 * Property 10: User List Ordering
 *
 * Feature: project-website
 */

import { assertEquals, assertRejects } from "https://deno.land/std@0.224.0/assert/mod.ts";
import fc from "fast-check";
import { initDatabase } from "../../db/mod.ts";
import { createUserService } from "../../services/users.ts";

// ─── Property 1: Email Uniqueness Invariant ───────────────────────────────────
/**
 * Property 1: Email Uniqueness Invariant
 *
 * For any set of User_Records in the Database, no two records share the same
 * email address. Generate random user creation sequences, verify the uniqueness
 * constraint is never violated (duplicate email throws or is rejected by
 * isEmailAvailable).
 *
 * Validates: Requirements 12.3, 14.3, 22.4, 24.3
 */
Deno.test("Property 1: Email Uniqueness Invariant - no two users can share the same email", async () => {
  await fc.assert(
    fc.asyncProperty(
      // Generate a random email address
      fc.string({ minLength: 3, maxLength: 20 })
        .filter((s) => /^[a-zA-Z0-9]+$/.test(s))
        .map((s) => `${s}@test.example.com`),
      // Generate two distinct user names
      fc.string({ minLength: 1, maxLength: 30 }).filter((s) => s.trim().length > 0),
      fc.string({ minLength: 1, maxLength: 30 }).filter((s) => s.trim().length > 0),
      async (email: string, name1: string, name2: string) => {
        const db = await initDatabase(":memory:");
        const userService = createUserService(db);

        // Get the "users" role ID for assignment
        const usersRole = db.prepare("SELECT id FROM roles WHERE name = 'users'").get() as { id: number };

        // Create the first user with this email — should succeed
        await userService.createUser({
          name: name1,
          email,
          password: "password123!",
          roleIds: [usersRole.id],
        });

        // Verify isEmailAvailable returns false for the same email
        assertEquals(
          userService.isEmailAvailable(email),
          false,
          "isEmailAvailable should return false for an existing email",
        );

        // Attempting to create a second user with the same email should throw
        // (SQLite UNIQUE constraint violation)
        let threw = false;
        try {
          await userService.createUser({
            name: name2,
            email,
            password: "password456!",
            roleIds: [usersRole.id],
          });
        } catch (_e) {
          threw = true;
        }

        assertEquals(
          threw,
          true,
          "Creating a user with a duplicate email should throw an error",
        );

        db.close();
      },
    ),
    { numRuns: 20 },
  );
});

// ─── Property 4: User CRUD Round-Trip ─────────────────────────────────────────
/**
 * Property 4: User CRUD Round-Trip
 *
 * Creating a user then reading it returns the same data. Updating a user then
 * reading it reflects the changes. Deleting a user then reading it returns null.
 * Generate random valid user data, perform create/read/update/delete cycles and
 * verify consistency.
 *
 * Validates: Requirements 12.1, 13.1, 14.1, 15.2
 */
Deno.test("Property 4: User CRUD Round-Trip - create/read/update/delete consistency", async () => {
  await fc.assert(
    fc.asyncProperty(
      // Generate random user data for creation
      fc.record({
        name: fc.string({ minLength: 1, maxLength: 50 }).filter((s) => s.trim().length > 0),
        email: fc.string({ minLength: 3, maxLength: 20 })
          .filter((s) => /^[a-zA-Z0-9]+$/.test(s))
          .map((s) => `${s}@crud-test.example.com`),
        password: fc.string({ minLength: 8, maxLength: 50 }),
      }),
      // Generate random update data
      fc.record({
        name: fc.string({ minLength: 1, maxLength: 50 }).filter((s) => s.trim().length > 0),
        email: fc.string({ minLength: 3, maxLength: 20 })
          .filter((s) => /^[a-zA-Z0-9]+$/.test(s))
          .map((s) => `${s}@crud-update.example.com`),
      }),
      async (createData, updateData) => {
        const db = await initDatabase(":memory:");
        const userService = createUserService(db);

        // Get the "users" role ID
        const usersRole = db.prepare("SELECT id FROM roles WHERE name = 'users'").get() as { id: number };

        // ─── CREATE then READ ─────────────────────────────────────────────
        const created = await userService.createUser({
          name: createData.name,
          email: createData.email,
          password: createData.password,
          roleIds: [usersRole.id],
        });

        const readAfterCreate = userService.getUser(created.id);
        assertEquals(readAfterCreate !== null, true, "User should exist after creation");
        assertEquals(readAfterCreate!.name, createData.name, "Read name should match created name");
        assertEquals(readAfterCreate!.email, createData.email, "Read email should match created email");
        assertEquals(readAfterCreate!.roles.length, 1, "User should have 1 role assigned");
        assertEquals(readAfterCreate!.roles[0]!.id, usersRole.id, "User should have the 'users' role");

        // ─── UPDATE then READ ─────────────────────────────────────────────
        const updated = userService.updateUser(created.id, {
          name: updateData.name,
          email: updateData.email,
          roleIds: [usersRole.id],
        });

        const readAfterUpdate = userService.getUser(created.id);
        assertEquals(readAfterUpdate !== null, true, "User should exist after update");
        assertEquals(readAfterUpdate!.name, updateData.name, "Read name should match updated name");
        assertEquals(readAfterUpdate!.email, updateData.email, "Read email should match updated email");

        // ─── DELETE then READ ─────────────────────────────────────────────
        userService.deleteUser(created.id);

        const readAfterDelete = userService.getUser(created.id);
        assertEquals(readAfterDelete, null, "User should be null after deletion");

        db.close();
      },
    ),
    { numRuns: 15 },
  );
});

// ─── Property 10: User List Ordering ──────────────────────────────────────────
/**
 * Property 10: User List Ordering
 *
 * The user list is always sorted in reverse chronological order by creation date.
 * For any two adjacent users in the list, the first user's createdAt is greater
 * than or equal to the second user's createdAt.
 *
 * Validates: Requirements 13.3
 */
Deno.test("Property 10: User List Ordering - users listed in reverse chronological order", async () => {
  await fc.assert(
    fc.asyncProperty(
      // Generate a random number of users to create (2-8)
      fc.integer({ min: 2, max: 8 }),
      async (numUsers: number) => {
        const db = await initDatabase(":memory:");
        const userService = createUserService(db);

        // Get the "users" role ID
        const usersRole = db.prepare("SELECT id FROM roles WHERE name = 'users'").get() as { id: number };

        // Create multiple users with distinct timestamps
        // We insert directly with explicit created_at to control ordering
        for (let i = 0; i < numUsers; i++) {
          // Use varying timestamps to ensure ordering is tested
          const timestamp = new Date(2025, 0, 1, 0, 0, i).toISOString().replace("T", " ").slice(0, 19);
          const email = `user-${i}-${crypto.randomUUID().slice(0, 8)}@order-test.com`;

          db.exec(
            "INSERT INTO users (name, email, password_hash, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
            [`User ${i}`, email, "hash_placeholder", timestamp, timestamp],
          );
        }

        // Retrieve the user list
        const users = userService.listUsers();

        // Verify ordering: for any two adjacent users, the first has createdAt >= second
        for (let i = 0; i < users.length - 1; i++) {
          const current = users[i]!;
          const next = users[i + 1]!;
          const currentDate = new Date(current.createdAt).getTime();
          const nextDate = new Date(next.createdAt).getTime();

          assertEquals(
            currentDate >= nextDate,
            true,
            `User list ordering violated: user at index ${i} (createdAt: ${current.createdAt}) should be >= user at index ${i + 1} (createdAt: ${next.createdAt})`,
          );
        }

        db.close();
      },
    ),
    { numRuns: 30 },
  );
});

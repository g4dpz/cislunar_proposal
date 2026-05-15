/**
 * Property-Based Tests for Role Service
 *
 * Property 2: Role Name Uniqueness Invariant
 * Property 5: Role Deletion Constraint
 * Property 8: User-Role Assignment Consistency
 *
 * Feature: project-website
 */

import { assertEquals } from "https://deno.land/std@0.224.0/assert/mod.ts";
import fc from "fast-check";
import { initDatabase } from "../../db/mod.ts";
import { createRoleService } from "../../services/roles.ts";
import { createUserService } from "../../services/users.ts";

// ─── Property 2: Role Name Uniqueness Invariant ───────────────────────────────
/**
 * Property 2: Role Name Uniqueness Invariant
 *
 * For any set of Role_Records in the Database, no two records share the same
 * name. Generate random role creation sequences, verify the uniqueness constraint
 * is never violated (duplicate name throws or is rejected by isNameAvailable).
 *
 * Validates: Requirements 16.2, 18.2
 */
Deno.test("Property 2: Role Name Uniqueness Invariant - no two roles can share the same name", async () => {
  await fc.assert(
    fc.asyncProperty(
      // Generate a random role name (alphanumeric to avoid SQL issues)
      fc.string({ minLength: 1, maxLength: 30 })
        .filter((s) => /^[a-zA-Z][a-zA-Z0-9_ -]*$/.test(s))
        .filter((s) => s !== "admin" && s !== "users"), // Avoid collisions with seeded roles
      // Generate two distinct descriptions
      fc.string({ minLength: 0, maxLength: 100 }),
      fc.string({ minLength: 0, maxLength: 100 }),
      async (roleName: string, desc1: string, desc2: string) => {
        const db = await initDatabase(":memory:");
        const roleService = createRoleService(db);

        // Create the first role with this name — should succeed
        roleService.createRole({ name: roleName, description: desc1 });

        // Verify isNameAvailable returns false for the same name
        assertEquals(
          roleService.isNameAvailable(roleName),
          false,
          "isNameAvailable should return false for an existing role name",
        );

        // Attempting to create a second role with the same name should throw
        // (SQLite UNIQUE constraint violation)
        let threw = false;
        try {
          roleService.createRole({ name: roleName, description: desc2 });
        } catch (_e) {
          threw = true;
        }

        assertEquals(
          threw,
          true,
          "Creating a role with a duplicate name should throw an error",
        );

        db.close();
      },
    ),
    { numRuns: 25 },
  );
});

// ─── Property 5: Role Deletion Constraint ─────────────────────────────────────
/**
 * Property 5: Role Deletion Constraint
 *
 * A role with assigned users cannot be deleted (operation fails with
 * success: false). A role with zero assigned users can always be deleted
 * successfully.
 *
 * Generate roles with and without user assignments, attempt deletion, verify
 * constraint enforcement.
 *
 * Validates: Requirements 19.2, 19.3
 */
Deno.test("Property 5: Role Deletion Constraint - roles with users cannot be deleted", async () => {
  await fc.assert(
    fc.asyncProperty(
      // Generate a unique role name
      fc.string({ minLength: 1, maxLength: 20 })
        .filter((s) => /^[a-zA-Z][a-zA-Z0-9_]*$/.test(s))
        .filter((s) => s !== "admin" && s !== "users"),
      // Whether to assign a user to this role
      fc.boolean(),
      async (roleName: string, assignUser: boolean) => {
        const db = await initDatabase(":memory:");
        const roleService = createRoleService(db);
        const userService = createUserService(db);

        // Create a test role
        const role = roleService.createRole({ name: roleName, description: "test role" });

        if (assignUser) {
          // Create a user and assign them to this role
          const uniqueEmail = `deltest-${crypto.randomUUID().slice(0, 8)}@test.com`;
          await userService.createUser({
            name: "Test User",
            email: uniqueEmail,
            password: "password123!",
            roleIds: [role.id],
          });

          // Attempt to delete the role — should fail
          const result = roleService.deleteRole(role.id);
          assertEquals(
            result.success,
            false,
            "Deleting a role with assigned users should fail",
          );
          assertEquals(
            typeof result.error,
            "string",
            "Failed deletion should include an error message",
          );

          // Verify the role still exists
          const roleAfter = roleService.getRole(role.id);
          assertEquals(
            roleAfter !== null,
            true,
            "Role should still exist after failed deletion",
          );
        } else {
          // No users assigned — deletion should succeed
          const result = roleService.deleteRole(role.id);
          assertEquals(
            result.success,
            true,
            "Deleting a role with no assigned users should succeed",
          );

          // Verify the role no longer exists
          const roleAfter = roleService.getRole(role.id);
          assertEquals(
            roleAfter,
            null,
            "Role should not exist after successful deletion",
          );
        }

        db.close();
      },
    ),
    { numRuns: 25 },
  );
});

// ─── Property 8: User-Role Assignment Consistency ─────────────────────────────
/**
 * Property 8: User-Role Assignment Consistency
 *
 * After assigning K roles to a user, reading that user's roles returns exactly
 * those K roles. The user_roles junction table contains exactly K entries for
 * that user.
 *
 * Generate random role subsets, assign to user via updateUser, read back and
 * verify exact match.
 *
 * Validates: Requirements 20.2, 20.3
 */
Deno.test("Property 8: User-Role Assignment Consistency - assigned roles match exactly on read-back", async () => {
  await fc.assert(
    fc.asyncProperty(
      // Generate a random number of extra roles to create (1-5)
      fc.integer({ min: 1, max: 5 }),
      // Generate a random seed for subset selection
      fc.integer({ min: 0, max: 1000 }),
      async (numExtraRoles: number, subsetSeed: number) => {
        const db = await initDatabase(":memory:");
        const roleService = createRoleService(db);
        const userService = createUserService(db);

        // Collect all available role IDs (seeded roles + newly created ones)
        const createdRoleIds: number[] = [];

        // Get the seeded roles
        const adminRole = db.prepare("SELECT id FROM roles WHERE name = 'admin'").get() as { id: number };
        const usersRole = db.prepare("SELECT id FROM roles WHERE name = 'users'").get() as { id: number };
        createdRoleIds.push(adminRole.id, usersRole.id);

        // Create additional roles
        for (let i = 0; i < numExtraRoles; i++) {
          const role = roleService.createRole({
            name: `test-role-${i}-${crypto.randomUUID().slice(0, 6)}`,
            description: `Test role ${i}`,
          });
          createdRoleIds.push(role.id);
        }

        // Create a user with the "users" role initially
        const uniqueEmail = `assign-${crypto.randomUUID().slice(0, 8)}@test.com`;
        const user = await userService.createUser({
          name: "Assignment Test User",
          email: uniqueEmail,
          password: "password123!",
          roleIds: [usersRole.id],
        });

        // Select a random subset of roles to assign (at least 1)
        // Use the subsetSeed to deterministically pick a subset
        const selectedRoleIds: number[] = [];
        for (let i = 0; i < createdRoleIds.length; i++) {
          // Use bit manipulation on the seed to decide inclusion
          if ((subsetSeed >> (i % 10)) & 1) {
            selectedRoleIds.push(createdRoleIds[i]!);
          }
        }
        // Ensure at least one role is selected
        if (selectedRoleIds.length === 0) {
          selectedRoleIds.push(createdRoleIds[0]!);
        }

        // Update the user with the selected roles
        userService.updateUser(user.id, {
          name: "Assignment Test User",
          email: uniqueEmail,
          roleIds: selectedRoleIds,
        });

        // Read back the user and verify roles match exactly
        const readUser = userService.getUser(user.id);
        assertEquals(
          readUser !== null,
          true,
          "User should exist after update",
        );

        const readRoleIds = readUser!.roles.map((r) => r.id).sort((a, b) => a - b);
        const expectedRoleIds = [...selectedRoleIds].sort((a, b) => a - b);

        assertEquals(
          readRoleIds.length,
          expectedRoleIds.length,
          `User should have exactly ${expectedRoleIds.length} roles, got ${readRoleIds.length}`,
        );

        assertEquals(
          readRoleIds,
          expectedRoleIds,
          "User's roles should match exactly the assigned role IDs",
        );

        // Also verify the junction table directly
        const junctionRows = db.prepare(
          "SELECT role_id FROM user_roles WHERE user_id = ? ORDER BY role_id",
        ).all(user.id) as Array<{ role_id: number }>;

        assertEquals(
          junctionRows.length,
          expectedRoleIds.length,
          `user_roles table should have exactly ${expectedRoleIds.length} entries for this user`,
        );

        const junctionRoleIds = junctionRows.map((r) => r.role_id);
        assertEquals(
          junctionRoleIds,
          expectedRoleIds,
          "Junction table role IDs should match the assigned roles",
        );

        db.close();
      },
    ),
    { numRuns: 20 },
  );
});

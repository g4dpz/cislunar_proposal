/**
 * Property-Based Tests for Auth Service
 *
 * Property 3: Password Hash Round-Trip
 * Property 6: Session Expiry
 *
 * Feature: project-website
 */

import { assertEquals } from "https://deno.land/std@0.224.0/assert/mod.ts";
import fc from "fast-check";
import { hash, compare } from "bcrypt";
import { initDatabase } from "../../db/mod.ts";
import { createAuthService } from "../../services/auth.ts";

// ─── Property 3: Password Hash Round-Trip ─────────────────────────────────────
/**
 * Property 3: Password Hash Round-Trip
 *
 * For any valid password P (≥8 chars), hashing P with bcrypt and then verifying
 * P against the hash always returns true. Verifying any other string Q ≠ P
 * against the hash always returns false.
 *
 * Validates: Requirements 22.2, 23.2
 */
Deno.test("Property 3: Password Hash Round-Trip - correct password verifies true, different password verifies false", async () => {
  await fc.assert(
    fc.asyncProperty(
      // Generate random passwords of length 8-50 (printable ASCII)
      fc.string({ minLength: 8, maxLength: 50 }).filter((s) => s.length >= 8),
      // Generate a different string to use as the wrong password
      fc.string({ minLength: 1, maxLength: 50 }),
      async (password: string, otherStr: string) => {
        // Hash the password
        const hashed = await hash(password);

        // Verifying the correct password against the hash returns true
        const correctResult = await compare(password, hashed);
        assertEquals(correctResult, true, `Expected correct password to verify as true`);

        // Verifying a different string against the hash returns false (only if Q ≠ P)
        if (otherStr !== password) {
          const wrongResult = await compare(otherStr, hashed);
          assertEquals(wrongResult, false, `Expected different string to verify as false`);
        }
      },
    ),
    { numRuns: 10 }, // bcrypt is slow, keep runs low
  );
});

// ─── Property 6: Session Expiry ──────────────────────────────────────────────
/**
 * Property 6: Session Expiry
 *
 * A session created at time T is valid for requests at time T + D where D < 1 hour,
 * and invalid for requests at time T + D where D ≥ 1 hour.
 *
 * Validates: Requirements 26.1, 26.2, 26.3
 */
Deno.test("Property 6: Session Expiry - session valid before 1 hour, invalid at or after 1 hour", async () => {
  await fc.assert(
    fc.asyncProperty(
      // Generate a random time offset in milliseconds
      // Range: 0ms to 3 hours (to test both valid and expired cases)
      fc.integer({ min: 0, max: 3 * 60 * 60 * 1000 }),
      async (offsetMs: number) => {
        const ONE_HOUR_MS = 60 * 60 * 1000;

        // Create a fresh in-memory database
        const db = await initDatabase(":memory:");
        const authService = createAuthService(db);

        // Register a user (creates a session with 1-hour expiry from now)
        const { sessionToken } = await authService.register(
          "Test User",
          `user-${crypto.randomUUID()}@test.com`,
          "password123",
        );

        // Manually set the session's expires_at to exactly 1 hour from a fixed reference time
        // We simulate "session created at time T" by setting expires_at = T + 1 hour
        // Then we set expires_at relative to "now + offset" to test the boundary
        const now = Date.now();
        // Set expires_at so that the session expires at (now + ONE_HOUR_MS - offsetMs)
        // This means: if offsetMs < ONE_HOUR_MS, session is still valid (expires in the future)
        //             if offsetMs >= ONE_HOUR_MS, session is expired (expires at or before now)
        const expiresAt = new Date(now + ONE_HOUR_MS - offsetMs).toISOString();
        db.exec("UPDATE sessions SET expires_at = ? WHERE id = ?", [expiresAt, sessionToken]);

        // Validate the session (checks against current time)
        const result = await authService.validateSession(sessionToken);

        if (offsetMs < ONE_HOUR_MS) {
          // Session should be valid (expires in the future)
          assertEquals(
            result !== null,
            true,
            `Session should be valid when offset (${offsetMs}ms) < 1 hour`,
          );
        } else {
          // Session should be expired (expires at or before now)
          assertEquals(
            result,
            null,
            `Session should be expired when offset (${offsetMs}ms) >= 1 hour`,
          );
        }

        db.close();
      },
    ),
    { numRuns: 50 },
  );
});

/**
 * Unit Tests for Access Control Middleware
 *
 * Tests that protected routes redirect unauthenticated users,
 * admin routes reject non-admin users, and guest-only routes
 * redirect authenticated users.
 *
 * Validates: Requirements 26.1, 28.1
 *
 * Feature: project-website
 */

import { assertEquals } from "https://deno.land/std@0.224.0/assert/mod.ts";
import { requireAuth, requireAdmin, guestOnly } from "../../middleware/auth.ts";

// ─── Mock Context Factory ─────────────────────────────────────────────────────

interface MockRole {
  id: number;
  name: string;
  description: string;
}

interface MockUser {
  id: number;
  name: string;
  email: string;
  roles: MockRole[];
  createdAt: string;
  updatedAt: string;
}

interface MockContextResult {
  redirectedTo: string | null;
  nextCalled: boolean;
}

function createMockContext(user: MockUser | null): {
  ctx: { state: { user: MockUser | null }; response: { redirect: (url: string) => void } };
  next: () => Promise<void>;
  result: MockContextResult;
} {
  const result: MockContextResult = {
    redirectedTo: null,
    nextCalled: false,
  };

  const ctx = {
    state: { user },
    response: {
      redirect(url: string) {
        result.redirectedTo = url;
      },
    },
  };

  const next = async () => {
    result.nextCalled = true;
  };

  return { ctx, next, result };
}

// ─── Test Users ───────────────────────────────────────────────────────────────

const adminUser: MockUser = {
  id: 1,
  name: "Admin",
  email: "admin@arthur.radio",
  roles: [{ id: 1, name: "admin", description: "Full system administration access" }],
  createdAt: new Date().toISOString(),
  updatedAt: new Date().toISOString(),
};

const regularUser: MockUser = {
  id: 2,
  name: "Regular User",
  email: "user@example.com",
  roles: [{ id: 2, name: "users", description: "Standard registered user access" }],
  createdAt: new Date().toISOString(),
  updatedAt: new Date().toISOString(),
};

// ─── requireAuth Tests ────────────────────────────────────────────────────────

Deno.test("requireAuth - redirects to /login when user is null", async () => {
  const middleware = requireAuth();
  const { ctx, next, result } = createMockContext(null);

  // deno-lint-ignore no-explicit-any
  await middleware(ctx as any, next);

  assertEquals(result.redirectedTo, "/login");
  assertEquals(result.nextCalled, false);
});

Deno.test("requireAuth - calls next when user is present", async () => {
  const middleware = requireAuth();
  const { ctx, next, result } = createMockContext(regularUser);

  // deno-lint-ignore no-explicit-any
  await middleware(ctx as any, next);

  assertEquals(result.nextCalled, true);
  assertEquals(result.redirectedTo, null);
});

// ─── requireAdmin Tests ───────────────────────────────────────────────────────

Deno.test("requireAdmin - redirects to /login when user is null", async () => {
  const middleware = requireAdmin();
  const { ctx, next, result } = createMockContext(null);

  // deno-lint-ignore no-explicit-any
  await middleware(ctx as any, next);

  assertEquals(result.redirectedTo, "/login");
  assertEquals(result.nextCalled, false);
});

Deno.test("requireAdmin - redirects to /login when user has no admin role", async () => {
  const middleware = requireAdmin();
  const { ctx, next, result } = createMockContext(regularUser);

  // deno-lint-ignore no-explicit-any
  await middleware(ctx as any, next);

  assertEquals(result.redirectedTo, "/login");
  assertEquals(result.nextCalled, false);
});

Deno.test("requireAdmin - calls next when user has admin role", async () => {
  const middleware = requireAdmin();
  const { ctx, next, result } = createMockContext(adminUser);

  // deno-lint-ignore no-explicit-any
  await middleware(ctx as any, next);

  assertEquals(result.nextCalled, true);
  assertEquals(result.redirectedTo, null);
});

// ─── guestOnly Tests ──────────────────────────────────────────────────────────

Deno.test("guestOnly - redirects to /profile when user is present", async () => {
  const middleware = guestOnly();
  const { ctx, next, result } = createMockContext(regularUser);

  // deno-lint-ignore no-explicit-any
  await middleware(ctx as any, next);

  assertEquals(result.redirectedTo, "/profile");
  assertEquals(result.nextCalled, false);
});

Deno.test("guestOnly - calls next when user is null", async () => {
  const middleware = guestOnly();
  const { ctx, next, result } = createMockContext(null);

  // deno-lint-ignore no-explicit-any
  await middleware(ctx as any, next);

  assertEquals(result.nextCalled, true);
  assertEquals(result.redirectedTo, null);
});

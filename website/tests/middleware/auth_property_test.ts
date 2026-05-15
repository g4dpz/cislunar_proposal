/**
 * Property-Based Tests for Auth Middleware
 *
 * Property 9: Access Control Enforcement
 *
 * Feature: project-website
 * Property 9: Access Control Enforcement
 *
 * For any protected route, a request without a valid session token always results
 * in a redirect to the login page (HTTP 302). A request with a valid session always
 * returns the page content (HTTP 200).
 *
 * Validates: Requirements 26.1, 28.1
 */

import { assertEquals } from "https://deno.land/std@0.224.0/assert/mod.ts";
import fc from "fast-check";
import { requireAuth, requireAdmin } from "../../middleware/auth.ts";

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

/**
 * Creates a mock Oak Context and Next function for testing middleware.
 * Returns the context, next function, and a result object to inspect behavior.
 */
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

// ─── Generators ───────────────────────────────────────────────────────────────

/** Generate a random role that is NOT "admin" */
const nonAdminRoleArb = fc.record({
  id: fc.integer({ min: 1, max: 1000 }),
  name: fc.string({ minLength: 1, maxLength: 30 }).filter((n) => n !== "admin"),
  description: fc.string({ minLength: 0, maxLength: 100 }),
});

/** Generate the admin role */
const adminRoleArb = fc.record({
  id: fc.integer({ min: 1, max: 1000 }),
  name: fc.constant("admin"),
  description: fc.string({ minLength: 0, maxLength: 100 }),
});

/** Generate a random authenticated user (non-admin) */
const nonAdminUserArb = fc.record({
  id: fc.integer({ min: 1, max: 10000 }),
  name: fc.string({ minLength: 1, maxLength: 50 }),
  email: fc.emailAddress(),
  roles: fc.array(nonAdminRoleArb, { minLength: 0, maxLength: 5 }),
  createdAt: fc.constant(new Date().toISOString()),
  updatedAt: fc.constant(new Date().toISOString()),
});

/** Generate a random authenticated user WITH admin role */
const adminUserArb = fc.record({
  id: fc.integer({ min: 1, max: 10000 }),
  name: fc.string({ minLength: 1, maxLength: 50 }),
  email: fc.emailAddress(),
  roles: fc.tuple(adminRoleArb, fc.array(nonAdminRoleArb, { minLength: 0, maxLength: 3 }))
    .map(([admin, others]) => [admin, ...others]),
  createdAt: fc.constant(new Date().toISOString()),
  updatedAt: fc.constant(new Date().toISOString()),
});

// ─── Property 9: Access Control Enforcement ───────────────────────────────────

/**
 * Property 9a: requireAuth - Unauthenticated requests redirect to /login
 *
 * For any request without a valid session (user is null), the requireAuth
 * middleware always redirects to /login and does NOT call next().
 *
 * Validates: Requirements 26.1, 28.1
 */
Deno.test("Property 9: Access Control Enforcement - unauthenticated requests always redirect to /login (requireAuth)", async () => {
  const middleware = requireAuth();

  await fc.assert(
    fc.asyncProperty(
      // Generate random "request context" scenarios — user is always null (unauthenticated)
      fc.constant(null),
      async (_nullUser) => {
        const { ctx, next, result } = createMockContext(null);

        // deno-lint-ignore no-explicit-any
        await middleware(ctx as any, next);

        // Must redirect to /login
        assertEquals(result.redirectedTo, "/login", "Unauthenticated request must redirect to /login");
        // Must NOT call next (request should not proceed)
        assertEquals(result.nextCalled, false, "Unauthenticated request must not call next()");
      },
    ),
    { numRuns: 50 },
  );
});

/**
 * Property 9b: requireAuth - Authenticated requests proceed (call next)
 *
 * For any request with a valid session (user is not null), the requireAuth
 * middleware always calls next() and does NOT redirect.
 *
 * Validates: Requirements 26.1, 28.1
 */
Deno.test("Property 9: Access Control Enforcement - authenticated requests always proceed (requireAuth)", async () => {
  const middleware = requireAuth();

  await fc.assert(
    fc.asyncProperty(
      // Generate random authenticated users (admin or non-admin)
      fc.oneof(nonAdminUserArb, adminUserArb),
      async (user) => {
        const { ctx, next, result } = createMockContext(user);

        // deno-lint-ignore no-explicit-any
        await middleware(ctx as any, next);

        // Must call next (request proceeds)
        assertEquals(result.nextCalled, true, "Authenticated request must call next()");
        // Must NOT redirect
        assertEquals(result.redirectedTo, null, "Authenticated request must not redirect");
      },
    ),
    { numRuns: 100 },
  );
});

/**
 * Property 9c: requireAdmin - Unauthenticated requests redirect to /login
 *
 * For any request without a valid session (user is null), the requireAdmin
 * middleware always redirects to /login and does NOT call next().
 *
 * Validates: Requirements 26.1, 28.1
 */
Deno.test("Property 9: Access Control Enforcement - unauthenticated requests always redirect to /login (requireAdmin)", async () => {
  const middleware = requireAdmin();

  await fc.assert(
    fc.asyncProperty(
      fc.constant(null),
      async (_nullUser) => {
        const { ctx, next, result } = createMockContext(null);

        // deno-lint-ignore no-explicit-any
        await middleware(ctx as any, next);

        // Must redirect to /login
        assertEquals(result.redirectedTo, "/login", "Unauthenticated request must redirect to /login");
        // Must NOT call next
        assertEquals(result.nextCalled, false, "Unauthenticated request must not call next()");
      },
    ),
    { numRuns: 50 },
  );
});

/**
 * Property 9d: requireAdmin - Authenticated non-admin requests redirect to /login
 *
 * For any request with a valid session but WITHOUT the "admin" role, the
 * requireAdmin middleware always redirects to /login and does NOT call next().
 *
 * Validates: Requirements 26.1, 28.1
 */
Deno.test("Property 9: Access Control Enforcement - authenticated non-admin requests redirect to /login (requireAdmin)", async () => {
  const middleware = requireAdmin();

  await fc.assert(
    fc.asyncProperty(
      nonAdminUserArb,
      async (user) => {
        const { ctx, next, result } = createMockContext(user);

        // deno-lint-ignore no-explicit-any
        await middleware(ctx as any, next);

        // Must redirect to /login (user lacks admin role)
        assertEquals(result.redirectedTo, "/login", "Non-admin authenticated request must redirect to /login");
        // Must NOT call next
        assertEquals(result.nextCalled, false, "Non-admin authenticated request must not call next()");
      },
    ),
    { numRuns: 100 },
  );
});

/**
 * Property 9e: requireAdmin - Authenticated admin requests proceed (call next)
 *
 * For any request with a valid session AND the "admin" role, the requireAdmin
 * middleware always calls next() and does NOT redirect.
 *
 * Validates: Requirements 26.1, 28.1
 */
Deno.test("Property 9: Access Control Enforcement - authenticated admin requests always proceed (requireAdmin)", async () => {
  const middleware = requireAdmin();

  await fc.assert(
    fc.asyncProperty(
      adminUserArb,
      async (user) => {
        const { ctx, next, result } = createMockContext(user);

        // deno-lint-ignore no-explicit-any
        await middleware(ctx as any, next);

        // Must call next (admin request proceeds)
        assertEquals(result.nextCalled, true, "Admin request must call next()");
        // Must NOT redirect
        assertEquals(result.redirectedTo, null, "Admin request must not redirect");
      },
    ),
    { numRuns: 100 },
  );
});

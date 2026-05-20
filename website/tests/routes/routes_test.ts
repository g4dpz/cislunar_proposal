// tests/routes/routes_test.ts — Integration tests for full page rendering
//
// Tests each route via Oak's app.handle() to verify complete HTML responses,
// correct status codes, content types, and security headers.

import { assert, assertEquals } from "https://deno.land/std@0.224.0/assert/mod.ts";
import { Application } from "@oak/oak";
import { initHandlebars } from "../../views/engine.ts";
import { createRouter } from "../../routes/mod.ts";
import { initDatabase } from "../../db/mod.ts";
import { securityHeaders } from "../../middleware/security.ts";
import { errorHandler } from "../../middleware/error.ts";
import { createAuthMiddleware } from "../../middleware/auth.ts";
import { createAuthService } from "../../services/auth.ts";
import { createUserService } from "../../services/users.ts";
import { createRoleService } from "../../services/roles.ts";
import { createOutreachService } from "../../services/outreach.ts";

// ─── Test Setup ───────────────────────────────────────────────────────────────

/**
 * Creates a fully configured Oak application for testing.
 * Initializes the Handlebars engine and an in-memory SQLite database.
 */
async function createTestApp(): Promise<Application> {
  const engine = await initHandlebars("./views");
  const db = await initDatabase(":memory:");
  const authService = createAuthService(db);
  const userService = createUserService(db);
  const roleService = createRoleService(db);
  const outreachService = createOutreachService(db);
  const router = createRouter(engine, db, authService, userService, roleService, outreachService);

  const app = new Application();

  // Middleware stack (same order as main.ts)
  app.use(securityHeaders);
  app.use(errorHandler);
  app.use(createAuthMiddleware(authService));
  app.use(router.routes());
  app.use(router.allowedMethods());

  return app;
}

// ─── Helper ───────────────────────────────────────────────────────────────────

async function fetchRoute(
  app: Application,
  path: string,
): Promise<Response> {
  const response = await app.handle(new Request(`http://localhost${path}`));
  assert(response !== undefined, `Expected a response for ${path}`);
  return response!;
}

// ─── Integration Tests: HTML Page Routes ──────────────────────────────────────

const HTML_ROUTES = [
  "/",
  "/roadmap",
  "/conops",
  "/contact",
  "/privacy",
];

// Routes that require authentication (redirect to /login when unauthenticated)
const AUTH_REQUIRED_ROUTES = [
  "/docs",
];

Deno.test("Integration: All HTML routes return 200 with text/html content-type", async () => {
  const app = await createTestApp();

  for (const route of HTML_ROUTES) {
    const response = await fetchRoute(app, route);
    assertEquals(
      response.status,
      200,
      `Route ${route} should return 200, got ${response.status}`,
    );

    const contentType = response.headers.get("content-type");
    assert(
      contentType !== null && contentType.includes("text/html"),
      `Route ${route} should return text/html content-type, got "${contentType}"`,
    );

    // Verify the response body is non-empty HTML
    const body = await response.text();
    assert(
      body.includes("<!DOCTYPE html>") || body.includes("<html"),
      `Route ${route} should return valid HTML content`,
    );
  }
});

Deno.test("Integration: Auth-required routes redirect to /login when unauthenticated", async () => {
  const app = await createTestApp();

  for (const route of AUTH_REQUIRED_ROUTES) {
    const response = await fetchRoute(app, route);
    assertEquals(
      response.status,
      302,
      `Route ${route} should return 302 redirect, got ${response.status}`,
    );

    const location = response.headers.get("location");
    assert(
      location !== null && location.includes("/login"),
      `Route ${route} should redirect to /login, got "${location}"`,
    );
    // Consume body to avoid resource leak
    await response.text();
  }
});

// ─── Integration Test: Sitemap ────────────────────────────────────────────────

Deno.test("Integration: /sitemap.xml returns 200 with application/xml content-type", async () => {
  const app = await createTestApp();
  const response = await fetchRoute(app, "/sitemap.xml");

  assertEquals(response.status, 200, "Sitemap should return 200");

  const contentType = response.headers.get("content-type");
  assert(
    contentType !== null && contentType.includes("application/xml"),
    `Sitemap should return application/xml content-type, got "${contentType}"`,
  );

  const body = await response.text();
  assert(
    body.includes("<urlset") && body.includes("<url>") && body.includes("<loc>"),
    "Sitemap should contain valid XML sitemap structure",
  );
});

// ─── Integration Test: 404 for Unknown Routes ─────────────────────────────────

Deno.test("Integration: Unknown route returns 404", async () => {
  const app = await createTestApp();
  const response = await fetchRoute(app, "/nonexistent-page");

  assertEquals(response.status, 404, "Unknown route should return 404");

  const body = await response.text();
  assert(
    body.includes("404"),
    "404 page should contain '404' text",
  );
});

// ─── Integration Test: Security Headers ───────────────────────────────────────

Deno.test("Integration: Security headers are present on all responses", async () => {
  const app = await createTestApp();

  const testRoutes = ["/", "/roadmap", "/sitemap.xml", "/nonexistent-page"];

  for (const route of testRoutes) {
    const response = await fetchRoute(app, route);

    // X-Content-Type-Options
    assertEquals(
      response.headers.get("x-content-type-options"),
      "nosniff",
      `Route ${route} should have X-Content-Type-Options: nosniff`,
    );

    // X-Frame-Options
    assertEquals(
      response.headers.get("x-frame-options"),
      "DENY",
      `Route ${route} should have X-Frame-Options: DENY`,
    );

    // Referrer-Policy
    assertEquals(
      response.headers.get("referrer-policy"),
      "strict-origin-when-cross-origin",
      `Route ${route} should have Referrer-Policy: strict-origin-when-cross-origin`,
    );

    // Content-Security-Policy
    const csp = response.headers.get("content-security-policy");
    assert(
      csp !== null && csp.includes("default-src"),
      `Route ${route} should have Content-Security-Policy header with default-src`,
    );

    // Strict-Transport-Security
    const hsts = response.headers.get("strict-transport-security");
    assert(
      hsts !== null && hsts.includes("max-age="),
      `Route ${route} should have Strict-Transport-Security header`,
    );

    // Consume body to avoid resource leaks
    await response.text();
  }
});

// ─── Integration Test: Internal Navigation Links ──────────────────────────────

Deno.test("Integration: All internal navigation links point to valid routes", async () => {
  const app = await createTestApp();
  const response = await fetchRoute(app, "/");
  const html = await response.text();

  // Extract all internal href links from the navigation
  const hrefMatches = html.matchAll(/href="(\/[^"]*?)"/g);
  const internalLinks = new Set<string>();

  for (const match of hrefMatches) {
    const href = match[1]!;
    // Skip static asset links (css, js, images)
    if (href.startsWith("/css/") || href.startsWith("/js/") || href.startsWith("/images/")) {
      continue;
    }
    internalLinks.add(href);
  }

  // Verify each internal link resolves to a valid route (not 404)
  for (const link of internalLinks) {
    const linkResponse = await fetchRoute(app, link);
    assert(
      linkResponse.status !== 404,
      `Internal link "${link}" should not return 404`,
    );
    await linkResponse.text();
  }
});

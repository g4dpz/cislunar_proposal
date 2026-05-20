// tests/routes/accessibility_test.ts — Accessibility verification tests
//
// Verifies semantic HTML structure, image alt attributes, keyboard-focusable
// interactive elements, and skip-link presence across all rendered pages.

import { assert } from "https://deno.land/std@0.224.0/assert/mod.ts";
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

async function createTestApp(): Promise<Application> {
  const engine = await initHandlebars("./views");
  const db = await initDatabase(":memory:");
  const authService = createAuthService(db);
  const userService = createUserService(db);
  const roleService = createRoleService(db);
  const outreachService = createOutreachService(db);
  const router = createRouter(engine, db, authService, userService, roleService, outreachService);

  const app = new Application();
  app.use(securityHeaders);
  app.use(errorHandler);
  app.use(createAuthMiddleware(authService));
  app.use(router.routes());
  app.use(router.allowedMethods());

  return app;
}

async function fetchRoute(app: Application, path: string): Promise<Response> {
  const response = await app.handle(new Request(`http://localhost${path}`));
  assert(response !== undefined, `Expected a response for ${path}`);
  return response!;
}

const ALL_ROUTES = [
  "/",
  "/roadmap",
  "/conops",
  "/contact",
  "/privacy",
];

// ─── Accessibility: Semantic HTML Structure ───────────────────────────────────

Deno.test("Accessibility: All pages contain semantic HTML structure (header, nav, main, footer)", async () => {
  const app = await createTestApp();

  for (const route of ALL_ROUTES) {
    const response = await fetchRoute(app, route);
    const html = await response.text();

    assert(
      /<header[\s>]/.test(html),
      `Route ${route} should contain a <header> element`,
    );

    assert(
      /<nav[\s>]/.test(html),
      `Route ${route} should contain a <nav> element`,
    );

    assert(
      /<main[\s>]/.test(html),
      `Route ${route} should contain a <main> element`,
    );

    assert(
      /<footer[\s>]/.test(html),
      `Route ${route} should contain a <footer> element`,
    );
  }
});

// ─── Accessibility: Image Alt Attributes ──────────────────────────────────────

Deno.test("Accessibility: All non-decorative images have alt attributes", async () => {
  const app = await createTestApp();

  for (const route of ALL_ROUTES) {
    const response = await fetchRoute(app, route);
    const html = await response.text();

    // Find all <img> tags
    const imgTags = html.match(/<img\s[^>]*>/gi) || [];

    for (const imgTag of imgTags) {
      // Skip decorative images (role="presentation" or role="none")
      if (/role="(presentation|none)"/.test(imgTag)) {
        continue;
      }

      // Non-decorative images must have a non-empty alt attribute
      const altMatch = imgTag.match(/alt="([^"]*)"/);
      assert(
        altMatch !== null,
        `Route ${route}: Non-decorative image must have an alt attribute: ${imgTag}`,
      );
      assert(
        altMatch![1]!.trim().length > 0,
        `Route ${route}: Non-decorative image must have a non-empty alt attribute: ${imgTag}`,
      );
    }
  }
});

// ─── Accessibility: Keyboard-Focusable Interactive Elements ───────────────────

Deno.test("Accessibility: Links and buttons have proper attributes for keyboard navigation", async () => {
  const app = await createTestApp();

  for (const route of ALL_ROUTES) {
    const response = await fetchRoute(app, route);
    const html = await response.text();

    // Check all <a> tags have href attributes (makes them keyboard-focusable)
    const anchorTags = html.match(/<a\s[^>]*>/gi) || [];
    for (const anchor of anchorTags) {
      assert(
        /href=/.test(anchor),
        `Route ${route}: All <a> elements should have href attribute for keyboard focus: ${anchor}`,
      );
    }

    // Check all <button> tags have type attributes
    const buttonTags = html.match(/<button\s[^>]*>/gi) || [];
    for (const button of buttonTags) {
      assert(
        /type=/.test(button),
        `Route ${route}: All <button> elements should have type attribute: ${button}`,
      );
    }
  }
});

// ─── Accessibility: Skip-Link Presence ────────────────────────────────────────

Deno.test("Accessibility: Skip-link is present in the layout", async () => {
  const app = await createTestApp();

  for (const route of ALL_ROUTES) {
    const response = await fetchRoute(app, route);
    const html = await response.text();

    // Check for a skip-link (typically links to #main-content or similar)
    const hasSkipLink =
      /class="[^"]*skip[^"]*"/.test(html) ||
      /href="#main/.test(html) ||
      /href="#content/.test(html) ||
      /skip.to.(main|content)/i.test(html);

    assert(
      hasSkipLink,
      `Route ${route} should contain a skip-link for keyboard navigation`,
    );
  }
});

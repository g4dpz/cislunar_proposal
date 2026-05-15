// middleware/auth.ts — Session validation and access control middleware

import type { Context, Next, Middleware } from "@oak/oak";
import type { AuthService, UserWithRoles } from "../services/auth.ts";

// ─── Constants ────────────────────────────────────────────────────────────────

export const SESSION_COOKIE_NAME = "arthur_session";

// ─── Main Auth Middleware ─────────────────────────────────────────────────────

/**
 * Creates the main auth middleware that runs on ALL requests.
 * Extracts the session token from the HTTP-only cookie, validates it,
 * and attaches the authenticated user (or null) to ctx.state.user.
 */
export function createAuthMiddleware(authService: AuthService): Middleware {
  return async (ctx: Context, next: Next) => {
    const token = await ctx.cookies.get(SESSION_COOKIE_NAME);

    if (token) {
      const user = await authService.validateSession(token);
      ctx.state.user = user;
    } else {
      ctx.state.user = null;
    }

    await next();
  };
}

// ─── Route-Level Middleware Factories ─────────────────────────────────────────

/**
 * Returns middleware that requires a valid authenticated session.
 * Redirects to /login if ctx.state.user is null.
 */
export function requireAuth(): Middleware {
  return async (ctx: Context, next: Next) => {
    if (!ctx.state.user) {
      ctx.response.redirect("/login");
      return;
    }
    await next();
  };
}

/**
 * Returns middleware that requires the authenticated user to have the "admin" role.
 * Redirects to /login if user is not authenticated or lacks the admin role.
 */
export function requireAdmin(): Middleware {
  return async (ctx: Context, next: Next) => {
    const user = ctx.state.user as UserWithRoles | null;

    if (!user) {
      ctx.response.redirect("/login");
      return;
    }

    const hasAdmin = user.roles.some((role) => role.name === "admin");
    if (!hasAdmin) {
      ctx.response.redirect("/login");
      return;
    }

    await next();
  };
}

/**
 * Returns middleware that only allows unauthenticated visitors.
 * Redirects to /profile if ctx.state.user is already set (user is logged in).
 */
export function guestOnly(): Middleware {
  return async (ctx: Context, next: Next) => {
    if (ctx.state.user) {
      ctx.response.redirect("/profile");
      return;
    }
    await next();
  };
}

// ─── Cookie Helpers ───────────────────────────────────────────────────────────

/**
 * Sets the session cookie on the response with secure defaults.
 * HTTP-only, Secure, SameSite=Lax, path=/.
 */
export async function setSessionCookie(ctx: Context, token: string): Promise<void> {
  await ctx.cookies.set(SESSION_COOKIE_NAME, token, {
    httpOnly: true,
    secure: true,
    sameSite: "lax",
    path: "/",
    maxAge: 60 * 60, // 1 hour in seconds
  });
}

/**
 * Clears the session cookie from the browser.
 */
export async function clearSessionCookie(ctx: Context): Promise<void> {
  await ctx.cookies.delete(SESSION_COOKIE_NAME, {
    path: "/",
  });
}

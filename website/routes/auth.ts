// routes/auth.ts — Login, register, and logout route handlers

import type { Middleware, Context } from "@oak/oak";
import type { HandlebarsEngine } from "../views/engine.ts";
import { renderPage } from "../views/engine.ts";
import type { AuthService } from "../services/auth.ts";
import {
  setSessionCookie,
  clearSessionCookie,
  SESSION_COOKIE_NAME,
} from "../middleware/auth.ts";
import { siteContent } from "../content/data.ts";
import { getPageMeta } from "../content/seo.ts";
import type { PageData } from "../content/data.ts";
import { getTemplateUser } from "./helpers.ts";

// ─── Helpers ──────────────────────────────────────────────────────────────────

function buildLoginPageData(options?: {
  errors?: string[];
  email?: string;
  flash?: { success?: string; error?: string };
}): PageData {
  const meta = getPageMeta("login");
  const nav = siteContent.nav.map((item) => ({
    ...item,
    active: false,
  }));

  return {
    meta,
    nav,
    activeSection: "login",
    content: {
      email: options?.email ?? "",
    },
    collaborators: siteContent.overview.collaborators,
    currentYear: new Date().getFullYear(),
    ...(options?.errors ? { errors: options.errors } : {}),
    ...(options?.flash ? { flash: options.flash } : {}),
  } as PageData & { errors?: string[]; flash?: { success?: string; error?: string } };
}

function buildRegisterPageData(options?: {
  errors?: string[];
  name?: string;
  email?: string;
  flash?: { success?: string; error?: string };
}): PageData {
  const meta = getPageMeta("register");
  const nav = siteContent.nav.map((item) => ({
    ...item,
    active: false,
  }));

  return {
    meta,
    nav,
    activeSection: "register",
    content: {
      name: options?.name ?? "",
      email: options?.email ?? "",
    },
    collaborators: siteContent.overview.collaborators,
    currentYear: new Date().getFullYear(),
    ...(options?.errors ? { errors: options.errors } : {}),
    ...(options?.flash ? { flash: options.flash } : {}),
  } as PageData & { errors?: string[]; flash?: { success?: string; error?: string } };
}

// ─── Handler Factories ────────────────────────────────────────────────────────

/**
 * GET /login — Render the login form (guestOnly middleware applied externally).
 */
export function loginGetHandler(engine: HandlebarsEngine): Middleware {
  return (ctx: Context) => {
    const pageData = buildLoginPageData();
    pageData.user = getTemplateUser(ctx);
    const html = renderPage(engine, "login", pageData);
    ctx.response.body = html;
    ctx.response.type = "text/html";
  };
}

/**
 * POST /login — Validate credentials, create session, set cookie, redirect to profile.
 */
export function loginPostHandler(
  authService: AuthService,
  engine: HandlebarsEngine,
): Middleware {
  return async (ctx: Context) => {
    const body = ctx.request.body;
    const formData = await body.formData();

    const email = formData.get("email")?.toString().trim() ?? "";
    const password = formData.get("password")?.toString() ?? "";

    // Validate required fields
    const errors: string[] = [];
    if (!email) errors.push("Email is required");
    if (!password) errors.push("Password is required");

    if (errors.length > 0) {
      const pageData = buildLoginPageData({ errors, email });
      const html = renderPage(engine, "login", pageData);
      ctx.response.status = 400;
      ctx.response.body = html;
      ctx.response.type = "text/html";
      return;
    }

    // Attempt login
    const result = await authService.login(email, password);

    if (!result) {
      const pageData = buildLoginPageData({
        errors: ["Email or password is incorrect"],
        email,
      });
      const html = renderPage(engine, "login", pageData);
      ctx.response.status = 401;
      ctx.response.body = html;
      ctx.response.type = "text/html";
      return;
    }

    // Success — set session cookie and redirect to profile
    await setSessionCookie(ctx, result.sessionToken);
    ctx.response.redirect("/profile");
  };
}

/**
 * GET /register — Render the registration form (guestOnly middleware applied externally).
 */
export function registerGetHandler(engine: HandlebarsEngine): Middleware {
  return (ctx: Context) => {
    const pageData = buildRegisterPageData();
    pageData.user = getTemplateUser(ctx);
    const html = renderPage(engine, "register", pageData);
    ctx.response.body = html;
    ctx.response.type = "text/html";
  };
}

/**
 * POST /register — Validate input, create user, create session, set cookie, redirect to profile.
 */
export function registerPostHandler(
  authService: AuthService,
  engine: HandlebarsEngine,
): Middleware {
  return async (ctx: Context) => {
    const body = ctx.request.body;
    const formData = await body.formData();

    const name = formData.get("name")?.toString().trim() ?? "";
    const email = formData.get("email")?.toString().trim() ?? "";
    const password = formData.get("password")?.toString() ?? "";

    // Validate required fields
    const errors: string[] = [];
    if (!name) errors.push("Name is required");
    if (!email) errors.push("Email is required");
    if (!password) errors.push("Password is required");
    if (password && password.length < 8) {
      errors.push("Password must be at least 8 characters");
    }

    if (errors.length > 0) {
      const pageData = buildRegisterPageData({ errors, name, email });
      const html = renderPage(engine, "register", pageData);
      ctx.response.status = 400;
      ctx.response.body = html;
      ctx.response.type = "text/html";
      return;
    }

    // Attempt registration
    try {
      const result = await authService.register(name, email, password);

      // Success — set session cookie and redirect to profile
      await setSessionCookie(ctx, result.sessionToken);
      ctx.response.redirect("/profile");
    } catch (error) {
      const message =
        error instanceof Error ? error.message : "Registration failed";
      const pageData = buildRegisterPageData({
        errors: [message],
        name,
        email,
      });
      const html = renderPage(engine, "register", pageData);
      ctx.response.status = 400;
      ctx.response.body = html;
      ctx.response.type = "text/html";
    }
  };
}

/**
 * GET /logout — Invalidate session, clear cookie, redirect to homepage.
 */
export function logoutHandler(authService: AuthService): Middleware {
  return async (ctx: Context) => {
    const token = await ctx.cookies.get(SESSION_COOKIE_NAME);

    if (token) {
      await authService.logout(token);
      await clearSessionCookie(ctx);
    }

    ctx.response.redirect("/");
  };
}

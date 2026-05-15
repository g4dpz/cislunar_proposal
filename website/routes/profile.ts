// routes/profile.ts — Profile view, update, and password change route handlers

import type { Middleware, Context } from "@oak/oak";
import type { HandlebarsEngine } from "../views/engine.ts";
import { renderPage } from "../views/engine.ts";
import type { AuthService, UserWithRoles } from "../services/auth.ts";
import type { UserService } from "../services/users.ts";
import { siteContent } from "../content/data.ts";
import { getPageMeta } from "../content/seo.ts";
import type { PageData } from "../content/data.ts";

// ─── Helpers ──────────────────────────────────────────────────────────────────

function buildProfilePageData(
  user: UserWithRoles,
  options?: {
    errors?: string[];
    flash?: { success?: string; error?: string };
  },
): PageData & { errors?: string[]; flash?: { success?: string; error?: string } } {
  const meta = getPageMeta("profile");
  const nav = siteContent.nav.map((item) => ({
    ...item,
    active: false,
  }));

  const isAdmin = user.roles.some((role) => role.name === "admin");

  return {
    meta,
    nav,
    activeSection: "profile",
    content: {
      user,
    },
    collaborators: siteContent.overview.collaborators,
    currentYear: new Date().getFullYear(),
    user: {
      id: user.id,
      name: user.name,
      email: user.email,
      roles: user.roles,
      isAdmin,
    },
    ...(options?.errors ? { errors: options.errors } : {}),
    ...(options?.flash ? { flash: options.flash } : {}),
  };
}

// ─── Handler Factories ────────────────────────────────────────────────────────

/**
 * GET /profile — Render the user's profile page.
 * Requires requireAuth() middleware applied externally.
 */
export function profileGetHandler(engine: HandlebarsEngine): Middleware {
  return (ctx: Context) => {
    const user = ctx.state.user as UserWithRoles;
    const pageData = buildProfilePageData(user);
    const html = renderPage(engine, "profile", pageData);
    ctx.response.body = html;
    ctx.response.type = "text/html";
  };
}

/**
 * POST /profile — Update user name and email.
 * Requires requireAuth() middleware applied externally.
 */
export function profileUpdateHandler(
  userService: UserService,
  engine: HandlebarsEngine,
): Middleware {
  return async (ctx: Context) => {
    const user = ctx.state.user as UserWithRoles;
    const body = ctx.request.body;
    const formData = await body.formData();

    const name = formData.get("name")?.toString().trim() ?? "";
    const email = formData.get("email")?.toString().trim() ?? "";

    // Validate required fields
    const errors: string[] = [];
    if (!name) errors.push("Name is required");
    if (!email) errors.push("Email is required");

    if (errors.length > 0) {
      const pageData = buildProfilePageData(user, { errors });
      const html = renderPage(engine, "profile", pageData);
      ctx.response.status = 400;
      ctx.response.body = html;
      ctx.response.type = "text/html";
      return;
    }

    // Check email uniqueness (excluding current user)
    if (!userService.isEmailAvailable(email, user.id)) {
      const pageData = buildProfilePageData(user, {
        errors: ["Email is already in use by another account"],
      });
      const html = renderPage(engine, "profile", pageData);
      ctx.response.status = 400;
      ctx.response.body = html;
      ctx.response.type = "text/html";
      return;
    }

    // Update user — preserve existing role assignments
    const roleIds = user.roles.map((r) => r.id);
    userService.updateUser(user.id, { name, email, roleIds });

    // Re-read updated user for display
    const updatedUser = userService.getUser(user.id);
    if (!updatedUser) {
      ctx.response.redirect("/login");
      return;
    }

    const pageData = buildProfilePageData(updatedUser, {
      flash: { success: "Profile updated successfully" },
    });
    const html = renderPage(engine, "profile", pageData);
    ctx.response.body = html;
    ctx.response.type = "text/html";
  };
}

/**
 * POST /profile/password — Change the user's password.
 * Requires requireAuth() middleware applied externally.
 */
export function profilePasswordHandler(
  authService: AuthService,
  engine: HandlebarsEngine,
): Middleware {
  return async (ctx: Context) => {
    const user = ctx.state.user as UserWithRoles;
    const body = ctx.request.body;
    const formData = await body.formData();

    const currentPassword = formData.get("currentPassword")?.toString() ?? "";
    const newPassword = formData.get("newPassword")?.toString() ?? "";
    const confirmPassword = formData.get("confirmPassword")?.toString() ?? "";

    // Validate required fields
    const errors: string[] = [];
    if (!currentPassword) errors.push("Current password is required");
    if (!newPassword) errors.push("New password is required");
    if (!confirmPassword) errors.push("Confirm password is required");

    // Validate new password length
    if (newPassword && newPassword.length < 8) {
      errors.push("New password must be at least 8 characters");
    }

    // Validate passwords match
    if (newPassword && confirmPassword && newPassword !== confirmPassword) {
      errors.push("New password and confirmation do not match");
    }

    if (errors.length > 0) {
      const pageData = buildProfilePageData(user, { errors });
      const html = renderPage(engine, "profile", pageData);
      ctx.response.status = 400;
      ctx.response.body = html;
      ctx.response.type = "text/html";
      return;
    }

    // Attempt password change
    const success = await authService.changePassword(
      user.id,
      currentPassword,
      newPassword,
    );

    if (!success) {
      const pageData = buildProfilePageData(user, {
        errors: ["Current password is incorrect"],
      });
      const html = renderPage(engine, "profile", pageData);
      ctx.response.status = 400;
      ctx.response.body = html;
      ctx.response.type = "text/html";
      return;
    }

    const pageData = buildProfilePageData(user, {
      flash: { success: "Password changed successfully" },
    });
    const html = renderPage(engine, "profile", pageData);
    ctx.response.body = html;
    ctx.response.type = "text/html";
  };
}

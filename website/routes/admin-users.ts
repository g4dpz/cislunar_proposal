// routes/admin-users.ts — Admin user CRUD route handlers

import type { RouterContext } from "@oak/oak";
import type { HandlebarsEngine } from "../views/engine.ts";
import { renderPage } from "../views/engine.ts";
import type { UserService } from "../services/users.ts";
import type { RoleService } from "../services/roles.ts";
import { siteContent } from "../content/data.ts";
import { getPageMeta } from "../content/seo.ts";
import type { PageData } from "../content/data.ts";
import { getTemplateUser } from "./helpers.ts";

// ─── Helpers ──────────────────────────────────────────────────────────────────

function buildAdminPageData(
  activeSection: string,
  content: Record<string, unknown>,
  options?: {
    errors?: string[];
    flash?: { success?: string; error?: string };
    user?: { id: number; name: string; email: string; roles: Array<{ id: number; name: string; description: string }>; isAdmin?: boolean } | null;
  },
): PageData & { errors?: string[]; flash?: { success?: string; error?: string } } {
  const meta = getPageMeta("admin");
  const nav = siteContent.nav.map((item) => ({
    ...item,
    active: false,
  }));

  return {
    meta,
    nav,
    activeSection,
    content,
    collaborators: siteContent.overview.collaborators,
    currentYear: new Date().getFullYear(),
    user: options?.user ?? null,
    ...(options?.errors ? { errors: options.errors } : {}),
    ...(options?.flash ? { flash: options.flash } : {}),
  };
}

// ─── Handler Factories ────────────────────────────────────────────────────────

/**
 * GET /admin/users — List all users.
 * Requires requireAdmin() middleware applied externally.
 */
export function usersListHandler(
  userService: UserService,
  engine: HandlebarsEngine,
) {
  return (ctx: RouterContext<"/admin/users">) => {
    const users = userService.listUsers();

    const flash = ctx.state.flash as { success?: string; error?: string } | undefined;
    const pageData = buildAdminPageData("admin-users", { users }, { ...(flash ? { flash } : {}), user: getTemplateUser(ctx) });

    const html = renderPage(engine, "admin/users-list", pageData);
    ctx.response.body = html;
    ctx.response.type = "text/html";
  };
}

/**
 * GET /admin/users/:id — User detail view.
 * Requires requireAdmin() middleware applied externally.
 */
export function userDetailHandler(
  userService: UserService,
  engine: HandlebarsEngine,
) {
  return (ctx: RouterContext<"/admin/users/:id">) => {
    const id = Number(ctx.params.id);
    if (isNaN(id)) {
      ctx.response.status = 404;
      ctx.response.body = "Not found";
      return;
    }

    const user = userService.getUser(id);
    if (!user) {
      ctx.response.status = 404;
      ctx.response.body = "User not found";
      return;
    }

    const pageData = buildAdminPageData("admin-users", { user }, { user: getTemplateUser(ctx) });
    const html = renderPage(engine, "admin/users-detail", pageData);
    ctx.response.body = html;
    ctx.response.type = "text/html";
  };
}

/**
 * GET /admin/users/new — Create user form.
 * Requires requireAdmin() middleware applied externally.
 */
export function userCreateFormHandler(
  roleService: RoleService,
  engine: HandlebarsEngine,
) {
  return (ctx: RouterContext<"/admin/users/new">) => {
    const allRoles = roleService.listRoles();
    const roles = allRoles.map((r) => ({
      id: r.id,
      name: r.name,
      checked: false,
    }));

    const pageData = buildAdminPageData("admin-users", {
      isEdit: false,
      user: { name: "", email: "" },
      roles,
    }, { user: getTemplateUser(ctx) });

    const html = renderPage(engine, "admin/users-form", pageData);
    ctx.response.body = html;
    ctx.response.type = "text/html";
  };
}

/**
 * POST /admin/users — Create a new user.
 * Requires requireAdmin() middleware applied externally.
 */
export function userCreateHandler(
  userService: UserService,
  roleService: RoleService,
  engine: HandlebarsEngine,
) {
  return async (ctx: RouterContext<"/admin/users">) => {
    const body = ctx.request.body;
    const formData = await body.formData();

    const name = formData.get("name")?.toString().trim() ?? "";
    const email = formData.get("email")?.toString().trim() ?? "";
    const password = formData.get("password")?.toString() ?? "";
    const roleIds = formData.getAll("roleIds").map((v) => Number(v));

    // Validate required fields
    const errors: string[] = [];
    if (!name) errors.push("Name is required");
    if (!email) errors.push("Email is required");
    if (!password) errors.push("Password is required");
    if (password && password.length < 8) {
      errors.push("Password must be at least 8 characters");
    }

    if (errors.length > 0) {
      const allRoles = roleService.listRoles();
      const roles = allRoles.map((r) => ({
        id: r.id,
        name: r.name,
        checked: roleIds.includes(r.id),
      }));

      const pageData = buildAdminPageData(
        "admin-users",
        { isEdit: false, user: { name, email }, roles },
        { errors, user: getTemplateUser(ctx) },
      );
      const html = renderPage(engine, "admin/users-form", pageData);
      ctx.response.status = 400;
      ctx.response.body = html;
      ctx.response.type = "text/html";
      return;
    }

    // Check email availability
    if (!userService.isEmailAvailable(email)) {
      const allRoles = roleService.listRoles();
      const roles = allRoles.map((r) => ({
        id: r.id,
        name: r.name,
        checked: roleIds.includes(r.id),
      }));

      const pageData = buildAdminPageData(
        "admin-users",
        { isEdit: false, user: { name, email }, roles },
        { errors: ["Email is already in use"], user: getTemplateUser(ctx) },
      );
      const html = renderPage(engine, "admin/users-form", pageData);
      ctx.response.status = 400;
      ctx.response.body = html;
      ctx.response.type = "text/html";
      return;
    }

    // Create user
    await userService.createUser({ name, email, password, roleIds });

    // Redirect to user list with success flash
    ctx.state.flash = { success: "User created successfully" };
    ctx.response.redirect("/admin/users");
  };
}

/**
 * GET /admin/users/:id/edit — Edit user form.
 * Requires requireAdmin() middleware applied externally.
 */
export function userEditFormHandler(
  userService: UserService,
  roleService: RoleService,
  engine: HandlebarsEngine,
) {
  return (ctx: RouterContext<"/admin/users/:id/edit">) => {
    const id = Number(ctx.params.id);
    if (isNaN(id)) {
      ctx.response.status = 404;
      ctx.response.body = "Not found";
      return;
    }

    const user = userService.getUser(id);
    if (!user) {
      ctx.response.status = 404;
      ctx.response.body = "User not found";
      return;
    }

    const userRoleIds = user.roles.map((r) => r.id);
    const allRoles = roleService.listRoles();
    const roles = allRoles.map((r) => ({
      id: r.id,
      name: r.name,
      checked: userRoleIds.includes(r.id),
    }));

    const pageData = buildAdminPageData("admin-users", {
      isEdit: true,
      user: { id: user.id, name: user.name, email: user.email },
      roles,
    }, { user: getTemplateUser(ctx) });

    const html = renderPage(engine, "admin/users-form", pageData);
    ctx.response.body = html;
    ctx.response.type = "text/html";
  };
}

/**
 * POST /admin/users/:id — Update an existing user.
 * Requires requireAdmin() middleware applied externally.
 */
export function userUpdateHandler(
  userService: UserService,
  roleService: RoleService,
  engine: HandlebarsEngine,
) {
  return async (ctx: RouterContext<"/admin/users/:id">) => {
    const id = Number(ctx.params.id);
    if (isNaN(id)) {
      ctx.response.status = 404;
      ctx.response.body = "Not found";
      return;
    }

    const body = ctx.request.body;
    const formData = await body.formData();

    const name = formData.get("name")?.toString().trim() ?? "";
    const email = formData.get("email")?.toString().trim() ?? "";
    const roleIds = formData.getAll("roleIds").map((v) => Number(v));

    // Validate required fields
    const errors: string[] = [];
    if (!name) errors.push("Name is required");
    if (!email) errors.push("Email is required");

    if (errors.length > 0) {
      const allRoles = roleService.listRoles();
      const roles = allRoles.map((r) => ({
        id: r.id,
        name: r.name,
        checked: roleIds.includes(r.id),
      }));

      const pageData = buildAdminPageData(
        "admin-users",
        { isEdit: true, user: { id, name, email }, roles },
        { errors, user: getTemplateUser(ctx) },
      );
      const html = renderPage(engine, "admin/users-form", pageData);
      ctx.response.status = 400;
      ctx.response.body = html;
      ctx.response.type = "text/html";
      return;
    }

    // Check email availability (excluding current user)
    if (!userService.isEmailAvailable(email, id)) {
      const allRoles = roleService.listRoles();
      const roles = allRoles.map((r) => ({
        id: r.id,
        name: r.name,
        checked: roleIds.includes(r.id),
      }));

      const pageData = buildAdminPageData(
        "admin-users",
        { isEdit: true, user: { id, name, email }, roles },
        { errors: ["Email is already in use by another account"], user: getTemplateUser(ctx) },
      );
      const html = renderPage(engine, "admin/users-form", pageData);
      ctx.response.status = 400;
      ctx.response.body = html;
      ctx.response.type = "text/html";
      return;
    }

    // Update user
    userService.updateUser(id, { name, email, roleIds });

    // Redirect to user list with success flash
    ctx.state.flash = { success: "User updated successfully" };
    ctx.response.redirect("/admin/users");
  };
}

/**
 * POST /admin/users/:id/delete — Delete a user.
 * Requires requireAdmin() middleware applied externally.
 */
export function userDeleteHandler(
  userService: UserService,
  _engine: HandlebarsEngine,
) {
  return (ctx: RouterContext<"/admin/users/:id/delete">) => {
    const id = Number(ctx.params.id);
    if (isNaN(id)) {
      ctx.response.status = 404;
      ctx.response.body = "Not found";
      return;
    }

    userService.deleteUser(id);

    // Redirect to user list with success flash
    ctx.state.flash = { success: "User deleted successfully" };
    ctx.response.redirect("/admin/users");
  };
}

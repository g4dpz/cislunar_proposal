// routes/admin-roles.ts — Admin role CRUD route handlers

import type { RouterContext } from "@oak/oak";
import type { HandlebarsEngine } from "../views/engine.ts";
import { renderPage } from "../views/engine.ts";
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
 * GET /admin/roles — List all roles.
 * Requires requireAdmin() middleware applied externally.
 */
export function rolesListHandler(
  roleService: RoleService,
  engine: HandlebarsEngine,
) {
  return (ctx: RouterContext<"/admin/roles">) => {
    const roles = roleService.listRoles();

    const flash = ctx.state.flash as { success?: string; error?: string } | undefined;
    const pageData = buildAdminPageData("admin-roles", { roles }, { ...(flash ? { flash } : {}), user: getTemplateUser(ctx) });

    const html = renderPage(engine, "admin/roles-list", pageData);
    ctx.response.body = html;
    ctx.response.type = "text/html";
  };
}

/**
 * GET /admin/roles/:id — Role detail view.
 * Requires requireAdmin() middleware applied externally.
 */
export function roleDetailHandler(
  roleService: RoleService,
  engine: HandlebarsEngine,
) {
  return (ctx: RouterContext<"/admin/roles/:id">) => {
    const id = Number(ctx.params.id);
    if (isNaN(id)) {
      ctx.response.status = 404;
      ctx.response.body = "Not found";
      return;
    }

    const role = roleService.getRole(id);
    if (!role) {
      ctx.response.status = 404;
      ctx.response.body = "Role not found";
      return;
    }

    const pageData = buildAdminPageData("admin-roles", { role }, { user: getTemplateUser(ctx) });
    const html = renderPage(engine, "admin/roles-detail", pageData);
    ctx.response.body = html;
    ctx.response.type = "text/html";
  };
}

/**
 * GET /admin/roles/new — Create role form.
 * Requires requireAdmin() middleware applied externally.
 */
export function roleCreateFormHandler(
  engine: HandlebarsEngine,
) {
  return (ctx: RouterContext<"/admin/roles/new">) => {
    const pageData = buildAdminPageData("admin-roles", {
      isEdit: false,
      role: { name: "", description: "" },
    }, { user: getTemplateUser(ctx) });

    const html = renderPage(engine, "admin/roles-form", pageData);
    ctx.response.body = html;
    ctx.response.type = "text/html";
  };
}

/**
 * POST /admin/roles — Create a new role.
 * Requires requireAdmin() middleware applied externally.
 */
export function roleCreateHandler(
  roleService: RoleService,
  engine: HandlebarsEngine,
) {
  return async (ctx: RouterContext<"/admin/roles">) => {
    const body = ctx.request.body;
    const formData = await body.formData();

    const name = formData.get("name")?.toString().trim() ?? "";
    const description = formData.get("description")?.toString().trim() ?? "";

    // Validate required fields
    const errors: string[] = [];
    if (!name) errors.push("Name is required");

    if (errors.length > 0) {
      const pageData = buildAdminPageData(
        "admin-roles",
        { isEdit: false, role: { name, description } },
        { errors, user: getTemplateUser(ctx) },
      );
      const html = renderPage(engine, "admin/roles-form", pageData);
      ctx.response.status = 400;
      ctx.response.body = html;
      ctx.response.type = "text/html";
      return;
    }

    // Check name availability
    if (!roleService.isNameAvailable(name)) {
      const pageData = buildAdminPageData(
        "admin-roles",
        { isEdit: false, role: { name, description } },
        { errors: ["Role name is already in use"], user: getTemplateUser(ctx) },
      );
      const html = renderPage(engine, "admin/roles-form", pageData);
      ctx.response.status = 400;
      ctx.response.body = html;
      ctx.response.type = "text/html";
      return;
    }

    // Create role
    roleService.createRole({ name, description });

    // Redirect to role list with success flash
    ctx.state.flash = { success: "Role created successfully" };
    ctx.response.redirect("/admin/roles");
  };
}

/**
 * GET /admin/roles/:id/edit — Edit role form.
 * Requires requireAdmin() middleware applied externally.
 */
export function roleEditFormHandler(
  roleService: RoleService,
  engine: HandlebarsEngine,
) {
  return (ctx: RouterContext<"/admin/roles/:id/edit">) => {
    const id = Number(ctx.params.id);
    if (isNaN(id)) {
      ctx.response.status = 404;
      ctx.response.body = "Not found";
      return;
    }

    const role = roleService.getRole(id);
    if (!role) {
      ctx.response.status = 404;
      ctx.response.body = "Role not found";
      return;
    }

    const pageData = buildAdminPageData("admin-roles", {
      isEdit: true,
      role: { id: role.id, name: role.name, description: role.description },
    }, { user: getTemplateUser(ctx) });

    const html = renderPage(engine, "admin/roles-form", pageData);
    ctx.response.body = html;
    ctx.response.type = "text/html";
  };
}

/**
 * POST /admin/roles/:id — Update an existing role.
 * Requires requireAdmin() middleware applied externally.
 */
export function roleUpdateHandler(
  roleService: RoleService,
  engine: HandlebarsEngine,
) {
  return async (ctx: RouterContext<"/admin/roles/:id">) => {
    const id = Number(ctx.params.id);
    if (isNaN(id)) {
      ctx.response.status = 404;
      ctx.response.body = "Not found";
      return;
    }

    const body = ctx.request.body;
    const formData = await body.formData();

    const name = formData.get("name")?.toString().trim() ?? "";
    const description = formData.get("description")?.toString().trim() ?? "";

    // Validate required fields
    const errors: string[] = [];
    if (!name) errors.push("Name is required");

    if (errors.length > 0) {
      const pageData = buildAdminPageData(
        "admin-roles",
        { isEdit: true, role: { id, name, description } },
        { errors, user: getTemplateUser(ctx) },
      );
      const html = renderPage(engine, "admin/roles-form", pageData);
      ctx.response.status = 400;
      ctx.response.body = html;
      ctx.response.type = "text/html";
      return;
    }

    // Check name availability (excluding current role)
    if (!roleService.isNameAvailable(name, id)) {
      const pageData = buildAdminPageData(
        "admin-roles",
        { isEdit: true, role: { id, name, description } },
        { errors: ["Role name is already in use by another role"], user: getTemplateUser(ctx) },
      );
      const html = renderPage(engine, "admin/roles-form", pageData);
      ctx.response.status = 400;
      ctx.response.body = html;
      ctx.response.type = "text/html";
      return;
    }

    // Update role
    roleService.updateRole(id, { name, description });

    // Redirect to role list with success flash
    ctx.state.flash = { success: "Role updated successfully" };
    ctx.response.redirect("/admin/roles");
  };
}

/**
 * POST /admin/roles/:id/delete — Delete a role.
 * Requires requireAdmin() middleware applied externally.
 */
export function roleDeleteHandler(
  roleService: RoleService,
  _engine: HandlebarsEngine,
) {
  return (ctx: RouterContext<"/admin/roles/:id/delete">) => {
    const id = Number(ctx.params.id);
    if (isNaN(id)) {
      ctx.response.status = 404;
      ctx.response.body = "Not found";
      return;
    }

    const result = roleService.deleteRole(id);

    if (!result.success) {
      // Role has assigned users — redirect back with error
      ctx.state.flash = { error: result.error ?? "Cannot delete role with assigned users" };
      ctx.response.redirect(`/admin/roles/${id}`);
      return;
    }

    // Redirect to role list with success flash
    ctx.state.flash = { success: "Role deleted successfully" };
    ctx.response.redirect("/admin/roles");
  };
}

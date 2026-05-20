// routes/outreach.ts — Admin outreach CRUD route handlers

import type { RouterContext } from "@oak/oak";
import type { HandlebarsEngine } from "../views/engine.ts";
import { renderPage } from "../views/engine.ts";
import type { OutreachService, ContactType, OutreachStatus } from "../services/outreach.ts";
import { siteContent } from "../content/data.ts";
import { getPageMeta } from "../content/seo.ts";
import type { PageData } from "../content/data.ts";
import { getTemplateUser } from "./helpers.ts";

// ─── Constants ────────────────────────────────────────────────────────────────

const VALID_CONTACT_TYPES: ContactType[] = [
  "facebook_group",
  "mailing_list",
  "club",
  "university",
  "individual",
  "cubesat_team",
  "organisation",
];

const VALID_STATUSES: OutreachStatus[] = [
  "identified",
  "contacted",
  "responded",
  "collaborating",
];

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
 * GET /admin/outreach — List all contacts with filters and stats.
 * Requires requireAdmin() middleware applied externally.
 */
export function outreachListHandler(
  outreachService: OutreachService,
  engine: HandlebarsEngine,
) {
  return (ctx: RouterContext<"/admin/outreach">) => {
    // Parse optional query params for filtering
    const statusParam = ctx.request.url.searchParams.get("status") as OutreachStatus | null;
    const typeParam = ctx.request.url.searchParams.get("type") as ContactType | null;

    const filters: { status?: OutreachStatus; contactType?: ContactType } = {};
    if (statusParam && VALID_STATUSES.includes(statusParam)) {
      filters.status = statusParam;
    }
    if (typeParam && VALID_CONTACT_TYPES.includes(typeParam)) {
      filters.contactType = typeParam;
    }

    const contacts = outreachService.listContacts(filters);
    const stats = outreachService.getStats();

    const flash = ctx.state.flash as { success?: string; error?: string } | undefined;
    const pageData = buildAdminPageData(
      "admin-outreach",
      {
        contacts,
        stats,
        filters: {
          status: statusParam ?? "",
          type: typeParam ?? "",
        },
        contactTypes: VALID_CONTACT_TYPES,
        statuses: VALID_STATUSES,
      },
      { ...(flash ? { flash } : {}), user: getTemplateUser(ctx) },
    );

    const html = renderPage(engine, "admin/outreach-list", pageData);
    ctx.response.body = html;
    ctx.response.type = "text/html";
  };
}

/**
 * GET /admin/outreach/new — Render create contact form.
 * Requires requireAdmin() middleware applied externally.
 */
export function outreachCreateFormHandler(
  _outreachService: OutreachService,
  engine: HandlebarsEngine,
) {
  return (ctx: RouterContext<"/admin/outreach/new">) => {
    const pageData = buildAdminPageData("admin-outreach", {
      isEdit: false,
      contact: {
        name: "",
        contactType: "",
        status: "identified",
        contactUrl: "",
        callsign: "",
        email: "",
        notes: "",
        contactedBy: "",
        contactedDate: "",
      },
      contactTypes: VALID_CONTACT_TYPES,
      statuses: VALID_STATUSES,
    }, { user: getTemplateUser(ctx) });

    const html = renderPage(engine, "admin/outreach-form", pageData);
    ctx.response.body = html;
    ctx.response.type = "text/html";
  };
}

/**
 * POST /admin/outreach — Create a new contact record.
 * Requires requireAdmin() middleware applied externally.
 */
export function outreachCreateHandler(
  outreachService: OutreachService,
  engine: HandlebarsEngine,
) {
  return async (ctx: RouterContext<"/admin/outreach">) => {
    const body = ctx.request.body;
    const formData = await body.formData();

    const name = formData.get("name")?.toString().trim() ?? "";
    const contactType = formData.get("contactType")?.toString().trim() ?? "";
    const status = formData.get("status")?.toString().trim() ?? "identified";
    const contactUrl = formData.get("contactUrl")?.toString().trim() ?? "";
    const callsign = formData.get("callsign")?.toString().trim() ?? "";
    const email = formData.get("email")?.toString().trim() ?? "";
    const notes = formData.get("notes")?.toString().trim() ?? "";
    const contactedBy = formData.get("contactedBy")?.toString().trim() ?? "";
    const contactedDate = formData.get("contactedDate")?.toString().trim() ?? "";

    // Validate required fields
    const errors: string[] = [];
    if (!name) errors.push("Name is required");
    if (!contactType) errors.push("Contact type is required");
    if (contactType && !VALID_CONTACT_TYPES.includes(contactType as ContactType)) {
      errors.push("Invalid contact type");
    }

    if (errors.length > 0) {
      const pageData = buildAdminPageData(
        "admin-outreach",
        {
          isEdit: false,
          contact: { name, contactType, status, contactUrl, callsign, email, notes, contactedBy, contactedDate },
          contactTypes: VALID_CONTACT_TYPES,
          statuses: VALID_STATUSES,
        },
        { errors, user: getTemplateUser(ctx) },
      );
      const html = renderPage(engine, "admin/outreach-form", pageData);
      ctx.response.status = 400;
      ctx.response.body = html;
      ctx.response.type = "text/html";
      return;
    }

    // Check for duplicates
    if (outreachService.isDuplicate(name, contactType as ContactType)) {
      const similar = outreachService.searchSimilar(name);
      const pageData = buildAdminPageData(
        "admin-outreach",
        {
          isEdit: false,
          contact: { name, contactType, status, contactUrl, callsign, email, notes, contactedBy, contactedDate },
          contactTypes: VALID_CONTACT_TYPES,
          statuses: VALID_STATUSES,
          duplicateWarning: `A contact with name "${name}" and type "${contactType}" already exists`,
          similarContacts: similar,
        },
        { errors: [`A contact with name "${name}" and type "${contactType}" already exists`], user: getTemplateUser(ctx) },
      );
      const html = renderPage(engine, "admin/outreach-form", pageData);
      ctx.response.status = 400;
      ctx.response.body = html;
      ctx.response.type = "text/html";
      return;
    }

    // Create the contact
    outreachService.createContact({
      name,
      contactType: contactType as ContactType,
      status: status as OutreachStatus,
      contactUrl: contactUrl || undefined,
      callsign: callsign || undefined,
      email: email || undefined,
      notes: notes || undefined,
      contactedBy: contactedBy || undefined,
      contactedDate: contactedDate || undefined,
    });

    ctx.state.flash = { success: "Contact created successfully" };
    ctx.response.redirect("/admin/outreach");
  };
}

/**
 * GET /admin/outreach/:id — Contact detail view.
 * Requires requireAdmin() middleware applied externally.
 */
export function outreachDetailHandler(
  outreachService: OutreachService,
  engine: HandlebarsEngine,
) {
  return (ctx: RouterContext<"/admin/outreach/:id">) => {
    const id = Number(ctx.params.id);
    if (isNaN(id)) {
      ctx.response.status = 404;
      ctx.response.body = "Not found";
      return;
    }

    const contact = outreachService.getContact(id);
    if (!contact) {
      ctx.response.status = 404;
      ctx.response.body = "Contact not found";
      return;
    }

    const pageData = buildAdminPageData("admin-outreach", { contact }, { user: getTemplateUser(ctx) });
    const html = renderPage(engine, "admin/outreach-detail", pageData);
    ctx.response.body = html;
    ctx.response.type = "text/html";
  };
}

/**
 * GET /admin/outreach/:id/edit — Render edit form with current values.
 * Requires requireAdmin() middleware applied externally.
 */
export function outreachEditFormHandler(
  outreachService: OutreachService,
  engine: HandlebarsEngine,
) {
  return (ctx: RouterContext<"/admin/outreach/:id/edit">) => {
    const id = Number(ctx.params.id);
    if (isNaN(id)) {
      ctx.response.status = 404;
      ctx.response.body = "Not found";
      return;
    }

    const contact = outreachService.getContact(id);
    if (!contact) {
      ctx.response.status = 404;
      ctx.response.body = "Contact not found";
      return;
    }

    const pageData = buildAdminPageData("admin-outreach", {
      isEdit: true,
      contact: {
        id: contact.id,
        name: contact.name,
        contactType: contact.contactType,
        status: contact.status,
        contactUrl: contact.contactUrl ?? "",
        callsign: contact.callsign ?? "",
        email: contact.email ?? "",
        notes: contact.notes ?? "",
        contactedBy: contact.contactedBy ?? "",
        contactedDate: contact.contactedDate ?? "",
      },
      contactTypes: VALID_CONTACT_TYPES,
      statuses: VALID_STATUSES,
    }, { user: getTemplateUser(ctx) });

    const html = renderPage(engine, "admin/outreach-form", pageData);
    ctx.response.body = html;
    ctx.response.type = "text/html";
  };
}

/**
 * POST /admin/outreach/:id — Update an existing contact record.
 * Requires requireAdmin() middleware applied externally.
 */
export function outreachUpdateHandler(
  outreachService: OutreachService,
  engine: HandlebarsEngine,
) {
  return async (ctx: RouterContext<"/admin/outreach/:id">) => {
    const id = Number(ctx.params.id);
    if (isNaN(id)) {
      ctx.response.status = 404;
      ctx.response.body = "Not found";
      return;
    }

    const body = ctx.request.body;
    const formData = await body.formData();

    const name = formData.get("name")?.toString().trim() ?? "";
    const contactType = formData.get("contactType")?.toString().trim() ?? "";
    const status = formData.get("status")?.toString().trim() ?? "identified";
    const contactUrl = formData.get("contactUrl")?.toString().trim() ?? "";
    const callsign = formData.get("callsign")?.toString().trim() ?? "";
    const email = formData.get("email")?.toString().trim() ?? "";
    const notes = formData.get("notes")?.toString().trim() ?? "";
    const contactedBy = formData.get("contactedBy")?.toString().trim() ?? "";
    const contactedDate = formData.get("contactedDate")?.toString().trim() ?? "";

    // Validate required fields
    const errors: string[] = [];
    if (!name) errors.push("Name is required");
    if (!contactType) errors.push("Contact type is required");
    if (contactType && !VALID_CONTACT_TYPES.includes(contactType as ContactType)) {
      errors.push("Invalid contact type");
    }

    if (errors.length > 0) {
      const pageData = buildAdminPageData(
        "admin-outreach",
        {
          isEdit: true,
          contact: { id, name, contactType, status, contactUrl, callsign, email, notes, contactedBy, contactedDate },
          contactTypes: VALID_CONTACT_TYPES,
          statuses: VALID_STATUSES,
        },
        { errors, user: getTemplateUser(ctx) },
      );
      const html = renderPage(engine, "admin/outreach-form", pageData);
      ctx.response.status = 400;
      ctx.response.body = html;
      ctx.response.type = "text/html";
      return;
    }

    // Check for duplicates (excluding current record)
    if (outreachService.isDuplicate(name, contactType as ContactType, id)) {
      const similar = outreachService.searchSimilar(name);
      const pageData = buildAdminPageData(
        "admin-outreach",
        {
          isEdit: true,
          contact: { id, name, contactType, status, contactUrl, callsign, email, notes, contactedBy, contactedDate },
          contactTypes: VALID_CONTACT_TYPES,
          statuses: VALID_STATUSES,
          duplicateWarning: `A contact with name "${name}" and type "${contactType}" already exists`,
          similarContacts: similar,
        },
        { errors: [`A contact with name "${name}" and type "${contactType}" already exists`], user: getTemplateUser(ctx) },
      );
      const html = renderPage(engine, "admin/outreach-form", pageData);
      ctx.response.status = 400;
      ctx.response.body = html;
      ctx.response.type = "text/html";
      return;
    }

    // Update the contact
    try {
      outreachService.updateContact(id, {
        name,
        contactType: contactType as ContactType,
        status: status as OutreachStatus,
        contactUrl: contactUrl || undefined,
        callsign: callsign || undefined,
        email: email || undefined,
        notes: notes || undefined,
        contactedBy: contactedBy || undefined,
        contactedDate: contactedDate || undefined,
      });
    } catch (_err) {
      ctx.response.status = 404;
      ctx.response.body = "Contact not found";
      return;
    }

    ctx.state.flash = { success: "Contact updated successfully" };
    ctx.response.redirect("/admin/outreach");
  };
}

/**
 * POST /admin/outreach/:id/delete — Delete a contact record.
 * Requires requireAdmin() middleware applied externally.
 */
export function outreachDeleteHandler(
  outreachService: OutreachService,
  _engine: HandlebarsEngine,
) {
  return (ctx: RouterContext<"/admin/outreach/:id/delete">) => {
    const id = Number(ctx.params.id);
    if (isNaN(id)) {
      ctx.response.status = 404;
      ctx.response.body = "Not found";
      return;
    }

    outreachService.deleteContact(id);

    ctx.state.flash = { success: "Contact deleted successfully" };
    ctx.response.redirect("/admin/outreach");
  };
}

/**
 * POST /admin/outreach/:id/advance — Advance contact status to next pipeline stage.
 * Requires requireAdmin() middleware applied externally.
 */
export function outreachAdvanceHandler(
  outreachService: OutreachService,
  _engine: HandlebarsEngine,
) {
  return (ctx: RouterContext<"/admin/outreach/:id/advance">) => {
    const id = Number(ctx.params.id);
    if (isNaN(id)) {
      ctx.response.status = 404;
      ctx.response.body = "Not found";
      return;
    }

    try {
      outreachService.advanceStatus(id);
      ctx.state.flash = { success: "Contact status advanced successfully" };
    } catch (err) {
      const message = err instanceof Error ? err.message : "Failed to advance status";
      ctx.state.flash = { error: message };
    }

    ctx.response.redirect("/admin/outreach");
  };
}

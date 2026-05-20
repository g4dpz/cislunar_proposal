// routes/collaborators.ts — Public collaborators page route handler

import type { RouterContext } from "@oak/oak";
import type { HandlebarsEngine } from "../views/engine.ts";
import { renderPage } from "../views/engine.ts";
import type { OutreachService, ContactType, ContactRecord } from "../services/outreach.ts";
import { siteContent } from "../content/data.ts";
import { getPageMeta } from "../content/seo.ts";
import type { PageData } from "../content/data.ts";
import { getTemplateUser } from "./helpers.ts";

// ─── Types ────────────────────────────────────────────────────────────────────

interface CollaboratorGroup {
  type: ContactType;
  label: string;
  contacts: ContactRecord[];
}

// ─── Constants ────────────────────────────────────────────────────────────────

const TYPE_LABELS: Record<ContactType, string> = {
  facebook_group: "Facebook Groups",
  mailing_list: "Mailing Lists",
  club: "Clubs",
  university: "Universities",
  individual: "Individuals",
  cubesat_team: "CubeSat Teams",
  organisation: "Organisations",
};

// ─── Handler Factory ──────────────────────────────────────────────────────────

/**
 * GET /collaborators — Public page showing collaborating contacts grouped by type.
 * No auth middleware required.
 */
export function collaboratorsHandler(
  outreachService: OutreachService,
  engine: HandlebarsEngine,
) {
  return (ctx: RouterContext<"/collaborators">) => {
    const collaborators = outreachService.getCollaborators();

    // Group collaborators by contact type
    const groupMap = new Map<ContactType, ContactRecord[]>();
    for (const contact of collaborators) {
      const existing = groupMap.get(contact.contactType) ?? [];
      existing.push(contact);
      groupMap.set(contact.contactType, existing);
    }

    // Convert to ordered array of groups (only include types that have contacts)
    const groups: CollaboratorGroup[] = [];
    for (const [type, contacts] of groupMap) {
      groups.push({
        type,
        label: TYPE_LABELS[type],
        contacts,
      });
    }

    const meta = getPageMeta("collaborators");
    const nav = siteContent.nav.map((item) => ({
      ...item,
      active: item.href === "/collaborators",
    }));

    const pageData: PageData = {
      meta,
      nav,
      activeSection: "collaborators",
      content: {
        groups,
        totalCount: collaborators.length,
      },
      collaborators: siteContent.overview.collaborators,
      currentYear: new Date().getFullYear(),
      user: getTemplateUser(ctx),
    };

    const html = renderPage(engine, "collaborators", pageData);
    ctx.response.body = html;
    ctx.response.type = "text/html";
  };
}

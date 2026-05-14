// routes/contact.ts — Contact page route handler (GET and POST)

import type { RouterContext } from "@oak/oak";
import type { HandlebarsEngine } from "../views/engine.ts";
import { renderPage } from "../views/engine.ts";
import { siteContent } from "../content/data.ts";
import { getPageMeta } from "../content/seo.ts";
import type { PageData } from "../content/data.ts";
import type { Database } from "@db/sqlite";
import { saveContactSubmission } from "../db/mod.ts";

function buildContactPageData(successMessage?: string): PageData {
  const meta = getPageMeta("contact");
  const nav = siteContent.nav.map((item) => ({
    ...item,
    active: item.href === "/contact",
  }));

  const pageData: PageData = {
    meta,
    nav,
    activeSection: "contact",
    content: {
      email: siteContent.contact.email,
      githubIssuesUrl: siteContent.contact.githubIssuesUrl,
      githubDiscussionsUrl: siteContent.contact.githubDiscussionsUrl,
      callsigns: siteContent.contact.callsigns,
      targetGroups: siteContent.contact.targetGroups,
      formFields: siteContent.contact.formFields,
      successMessage: successMessage ?? null,
    },
    collaborators: siteContent.contact.collaborators,
    currentYear: new Date().getFullYear(),
  };

  return pageData;
}

export function contactGetHandler(engine: HandlebarsEngine) {
  return (ctx: RouterContext<"/contact">) => {
    const successMessage = ctx.request.url.searchParams.get("success")
      ? "Thank you for your message! We will be in touch soon."
      : undefined;

    const pageData = buildContactPageData(successMessage);
    const html = renderPage(engine, "contact", pageData);
    ctx.response.body = html;
    ctx.response.type = "text/html";
  };
}

export function contactPostHandler(db: Database) {
  return async (ctx: RouterContext<"/contact">) => {
    const body = ctx.request.body;
    const formData = await body.formData();

    const name = formData.get("name")?.toString() ?? "";
    const callsignOrOrg = formData.get("callsign_or_org")?.toString() ?? "";
    const areaOfInterest = formData.get("area_of_interest")?.toString() ?? "";
    const message = formData.get("message")?.toString() ?? "";

    if (!name || !message) {
      ctx.response.status = 400;
      ctx.response.body = "Name and message are required.";
      return;
    }

    saveContactSubmission(db, {
      name,
      callsignOrOrg,
      areaOfInterest,
      message,
    });

    ctx.response.redirect("/contact?success=1");
  };
}

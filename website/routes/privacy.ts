// routes/privacy.ts — Privacy policy page route handler

import type { RouterContext } from "@oak/oak";
import type { HandlebarsEngine } from "../views/engine.ts";
import { renderPage } from "../views/engine.ts";
import { siteContent } from "../content/data.ts";
import { getPageMeta } from "../content/seo.ts";
import type { PageData } from "../content/data.ts";

export function privacyHandler(engine: HandlebarsEngine) {
  return (ctx: RouterContext<"/privacy">) => {
    const meta = getPageMeta("privacy");
    const nav = siteContent.nav.map((item) => ({
      ...item,
      active: item.href === "/privacy",
    }));

    const pageData: PageData = {
      meta,
      nav,
      activeSection: "privacy",
      content: {
        dataController: siteContent.privacy.dataController,
        contactEmail: siteContent.privacy.contactEmail,
        dataCollected: siteContent.privacy.dataCollected,
        legalBasis: siteContent.privacy.legalBasis,
        retentionPeriod: siteContent.privacy.retentionPeriod,
        cookiePolicy: siteContent.privacy.cookiePolicy,
      },
      collaborators: siteContent.overview.collaborators,
      currentYear: new Date().getFullYear(),
    };

    const html = renderPage(engine, "privacy", pageData);
    ctx.response.body = html;
    ctx.response.type = "text/html";
  };
}

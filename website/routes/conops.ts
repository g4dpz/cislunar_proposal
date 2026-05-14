// routes/conops.ts — Concept of Operations page route handler

import type { RouterContext } from "@oak/oak";
import type { HandlebarsEngine } from "../views/engine.ts";
import { renderPage } from "../views/engine.ts";
import { siteContent } from "../content/data.ts";
import { getPageMeta } from "../content/seo.ts";
import type { PageData } from "../content/data.ts";

export function conopsHandler(engine: HandlebarsEngine) {
  return (ctx: RouterContext<"/conops">) => {
    const meta = getPageMeta("conops");
    const nav = siteContent.nav.map((item) => ({
      ...item,
      active: item.href === "/conops",
    }));

    const pageData: PageData = {
      meta,
      nav,
      activeSection: "conops",
      content: {
        concept: siteContent.conops.concept,
        rfLinkTypes: siteContent.conops.rfLinkTypes,
        whyCommunityMatters: siteContent.conops.whyCommunityMatters,
        expectedOutcomes: siteContent.conops.expectedOutcomes,
        nasaReferences: siteContent.conops.nasaReferences,
      },
      collaborators: siteContent.overview.collaborators,
      currentYear: new Date().getFullYear(),
    };

    const html = renderPage(engine, "conops", pageData);
    ctx.response.body = html;
    ctx.response.type = "text/html";
  };
}

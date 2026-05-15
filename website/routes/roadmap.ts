// routes/roadmap.ts — Roadmap page route handler

import type { RouterContext } from "@oak/oak";
import type { HandlebarsEngine } from "../views/engine.ts";
import { renderPage } from "../views/engine.ts";
import { siteContent } from "../content/data.ts";
import { getPageMeta } from "../content/seo.ts";
import type { PageData } from "../content/data.ts";
import { getTemplateUser } from "./helpers.ts";

export function roadmapHandler(engine: HandlebarsEngine) {
  return (ctx: RouterContext<"/roadmap">) => {
    const meta = getPageMeta("roadmap");
    const nav = siteContent.nav.map((item) => ({
      ...item,
      active: item.href === "/roadmap",
    }));

    const pageData: PageData = {
      meta,
      nav,
      activeSection: "roadmap",
      content: {},
      collaborators: siteContent.overview.collaborators,
      currentYear: new Date().getFullYear(),
      user: getTemplateUser(ctx),
    };

    // Roadmap template accesses `roadmap` directly at the top level
    const dataWithRoadmap = {
      ...pageData,
      roadmap: siteContent.roadmap,
    };

    const html = renderPage(engine, "roadmap", dataWithRoadmap as unknown as PageData);
    ctx.response.body = html;
    ctx.response.type = "text/html";
  };
}

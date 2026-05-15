// routes/home.ts — Homepage route handler

import type { RouterContext } from "@oak/oak";
import type { HandlebarsEngine } from "../views/engine.ts";
import { renderPage } from "../views/engine.ts";
import { siteContent } from "../content/data.ts";
import { getPageMeta } from "../content/seo.ts";
import type { PageData } from "../content/data.ts";

export function homeHandler(engine: HandlebarsEngine) {
  return (ctx: RouterContext<"/">) => {
    const meta = getPageMeta("home");
    const nav = siteContent.nav.map((item) => ({
      ...item,
      active: item.href === "/",
    }));

    const pageData: PageData = {
      meta,
      nav,
      activeSection: "home",
      content: {
        title: siteContent.overview.title,
        tagline: siteContent.overview.tagline,
        missionSummary: siteContent.overview.missionSummary,
        missionSummary2: siteContent.overview.missionSummary2,
        features: siteContent.overview.features,
        protocolStack: siteContent.overview.protocolStack,
        license: siteContent.overview.license,
      },
      collaborators: siteContent.overview.collaborators,
      currentYear: new Date().getFullYear(),
    };

    const html = renderPage(engine, "home", pageData);
    ctx.response.body = html;
    ctx.response.type = "text/html";
  };
}

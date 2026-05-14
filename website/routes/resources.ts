// routes/resources.ts — Resources page route handler

import type { RouterContext } from "@oak/oak";
import type { HandlebarsEngine } from "../views/engine.ts";
import { renderPage } from "../views/engine.ts";
import { siteContent } from "../content/data.ts";
import { getPageMeta } from "../content/seo.ts";
import type { PageData } from "../content/data.ts";

export function resourcesHandler(engine: HandlebarsEngine) {
  return (ctx: RouterContext<"/resources">) => {
    const meta = getPageMeta("resources");
    const nav = siteContent.nav.map((item) => ({
      ...item,
      active: item.href === "/resources",
    }));

    const pageData: PageData = {
      meta,
      nav,
      activeSection: "resources",
      content: {},
      collaborators: siteContent.overview.collaborators,
      currentYear: new Date().getFullYear(),
    };

    // Resources template accesses `resources.categories` at top level
    const dataWithResources = {
      ...pageData,
      resources: siteContent.resources,
    };

    const html = renderPage(engine, "resources", dataWithResources as unknown as PageData);
    ctx.response.body = html;
    ctx.response.type = "text/html";
  };
}

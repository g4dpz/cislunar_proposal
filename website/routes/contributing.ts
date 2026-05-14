// routes/contributing.ts — Contributing page route handler

import type { RouterContext } from "@oak/oak";
import type { HandlebarsEngine } from "../views/engine.ts";
import { renderPage } from "../views/engine.ts";
import { siteContent } from "../content/data.ts";
import { getPageMeta } from "../content/seo.ts";
import type { PageData } from "../content/data.ts";

export function contributingHandler(engine: HandlebarsEngine) {
  return (ctx: RouterContext<"/contributing">) => {
    const meta = getPageMeta("contributing");
    const nav = siteContent.nav.map((item) => ({
      ...item,
      active: item.href === "/contributing",
    }));

    const pageData: PageData = {
      meta,
      nav,
      activeSection: "contributing",
      content: {
        areas: siteContent.contributing.areas,
        developmentSetup: siteContent.contributing.developmentSetup,
        license: siteContent.contributing.license,
      },
      collaborators: siteContent.overview.collaborators,
      currentYear: new Date().getFullYear(),
    };

    const html = renderPage(engine, "contributing", pageData);
    ctx.response.body = html;
    ctx.response.type = "text/html";
  };
}

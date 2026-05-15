// routes/docs.ts — Documentation page route handler

import type { RouterContext } from "@oak/oak";
import type { HandlebarsEngine } from "../views/engine.ts";
import { renderPage } from "../views/engine.ts";
import { siteContent } from "../content/data.ts";
import { getPageMeta } from "../content/seo.ts";
import type { PageData } from "../content/data.ts";
import { getTemplateUser } from "./helpers.ts";

export function docsHandler(engine: HandlebarsEngine) {
  return (ctx: RouterContext<"/docs">) => {
    const meta = getPageMeta("docs");
    const nav = siteContent.nav.map((item) => ({
      ...item,
      active: item.href === "/docs",
    }));

    const pageData: PageData = {
      meta,
      nav,
      activeSection: "docs",
      content: {
        phases: siteContent.documentation.phases,
        externalRefs: siteContent.documentation.externalRefs,
        packages: siteContent.documentation.packages,
      },
      collaborators: siteContent.overview.collaborators,
      currentYear: new Date().getFullYear(),
      user: getTemplateUser(ctx),
    };

    const html = renderPage(engine, "docs", pageData);
    ctx.response.body = html;
    ctx.response.type = "text/html";
  };
}

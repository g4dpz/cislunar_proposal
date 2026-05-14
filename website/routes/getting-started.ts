// routes/getting-started.ts — Getting Started page route handler

import type { RouterContext } from "@oak/oak";
import type { HandlebarsEngine } from "../views/engine.ts";
import { renderPage } from "../views/engine.ts";
import { siteContent } from "../content/data.ts";
import { getPageMeta } from "../content/seo.ts";
import type { PageData } from "../content/data.ts";

export function gettingStartedHandler(engine: HandlebarsEngine) {
  return (ctx: RouterContext<"/getting-started">) => {
    const meta = getPageMeta("getting-started");
    const nav = siteContent.nav.map((item) => ({
      ...item,
      active: item.href === "/getting-started",
    }));

    const pageData: PageData = {
      meta,
      nav,
      activeSection: "getting-started",
      content: {
        prerequisites: siteContent.gettingStarted.prerequisites,
        installation: siteContent.gettingStarted.installation,
        runNetwork: siteContent.gettingStarted.runNetwork,
        testConnectivity: siteContent.gettingStarted.testConnectivity,
      },
      collaborators: siteContent.overview.collaborators,
      currentYear: new Date().getFullYear(),
    };

    const html = renderPage(engine, "getting-started", pageData);
    ctx.response.body = html;
    ctx.response.type = "text/html";
  };
}

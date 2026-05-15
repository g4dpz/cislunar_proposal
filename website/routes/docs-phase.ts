// routes/docs-phase.ts — Dynamic route to render phase requirements from docs/

import type { RouterContext } from "@oak/oak";
import type { HandlebarsEngine } from "../views/engine.ts";
import { renderPage } from "../views/engine.ts";
import { siteContent } from "../content/data.ts";
import { getPageMeta } from "../content/seo.ts";
import type { PageData } from "../content/data.ts";
import { marked } from "marked";

/**
 * Serves rendered markdown requirements for a given phase.
 * Route: /docs/:phase/requirements
 *
 * Reads docs/<phase>/requirements.md from disk, converts to HTML,
 * and renders it within the site layout.
 */
export function docsPhaseHandler(engine: HandlebarsEngine) {
  return async (ctx: RouterContext<"/docs/:phase/requirements">) => {
    const phase = ctx.params.phase;

    // Validate the phase exists in our known phases
    const knownPhase = siteContent.documentation.phases.find((p) =>
      p.docsPath.includes(phase)
    );

    if (!knownPhase) {
      ctx.response.status = 404;
      ctx.response.body = "Phase not found";
      return;
    }

    // Read the requirements.md from the docs directory
    const docsPath = `../docs/${phase}/requirements.md`;
    let markdown: string;
    try {
      markdown = await Deno.readTextFile(docsPath);
    } catch (error) {
      if (error instanceof Deno.errors.NotFound) {
        ctx.response.status = 404;
        ctx.response.body = "Requirements document not found";
        return;
      }
      throw error;
    }

    // Convert markdown to HTML
    const htmlContent = await marked(markdown);

    const meta = getPageMeta("docs");
    const nav = siteContent.nav.map((item) => ({
      ...item,
      active: item.href === "/docs",
    }));

    const pageData: PageData = {
      meta: {
        ...meta,
        title: `${knownPhase.name} Requirements — RADIANT`,
      },
      nav,
      activeSection: "docs",
      content: {
        phaseName: knownPhase.name,
        htmlContent,
      },
      collaborators: siteContent.overview.collaborators,
      currentYear: new Date().getFullYear(),
    };

    const html = renderPage(engine, "docs-phase", pageData);
    ctx.response.body = html;
    ctx.response.type = "text/html";
  };
}

// routes/docs-architecture.ts — Route to serve architecture technical docs

import type { RouterContext } from "@oak/oak";
import type { HandlebarsEngine } from "../views/engine.ts";
import { renderPage } from "../views/engine.ts";
import { siteContent } from "../content/data.ts";
import { getPageMeta } from "../content/seo.ts";
import type { PageData } from "../content/data.ts";
import { marked } from "marked";
import { getTemplateUser } from "./helpers.ts";

/**
 * Serves rendered architecture markdown documents.
 * Route: /docs/architecture/:doc
 *
 * Reads docs/<filename>.md from disk, converts to HTML,
 * and renders it within the site layout.
 */
export function docsArchitectureHandler(engine: HandlebarsEngine) {
  return async (ctx: RouterContext<"/docs/architecture/:doc">) => {
    const doc = ctx.params.doc;

    // Validate against known architecture docs
    const knownDoc = siteContent.documentation.architectureDocs.find(
      (d) => d.slug === doc,
    );

    if (!knownDoc) {
      ctx.response.status = 404;
      ctx.response.body = "Document not found";
      return;
    }

    // Read the markdown file from the docs directory
    const docsPath = `../docs/${knownDoc.filename}`;
    let markdown: string;
    try {
      markdown = await Deno.readTextFile(docsPath);
    } catch (error) {
      if (error instanceof Deno.errors.NotFound) {
        ctx.response.status = 404;
        ctx.response.body = "Document not found";
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
        title: `${knownDoc.name} — RADIANT`,
      },
      nav,
      activeSection: "docs",
      content: {
        docName: knownDoc.name,
        htmlContent,
      },
      collaborators: siteContent.overview.collaborators,
      currentYear: new Date().getFullYear(),
      user: getTemplateUser(ctx),
    };

    const html = renderPage(engine, "docs-architecture", pageData);
    ctx.response.body = html;
    ctx.response.type = "text/html";
  };
}

// routes/mod.ts — Router setup and route registration

import { Router } from "@oak/oak";
import type { HandlebarsEngine } from "../views/engine.ts";
import type { Database } from "@db/sqlite";
import { homeHandler } from "./home.ts";
import { roadmapHandler } from "./roadmap.ts";
import { conopsHandler } from "./conops.ts";
import { docsHandler } from "./docs.ts";
import { docsPhaseHandler } from "./docs-phase.ts";
import { contactGetHandler, contactPostHandler } from "./contact.ts";
import { privacyHandler } from "./privacy.ts";
import { sitemapHandler } from "./sitemap.ts";

/**
 * Creates and configures the Oak router with all page routes.
 */
export function createRouter(engine: HandlebarsEngine, db: Database): Router {
  const router = new Router();

  // Page routes
  router.get("/", homeHandler(engine));
  router.get("/roadmap", roadmapHandler(engine));
  router.get("/conops", conopsHandler(engine));
  router.get("/docs", docsHandler(engine));
  router.get("/docs/:phase/requirements", docsPhaseHandler(engine));
  router.get("/contact", contactGetHandler(engine));
  router.post("/contact", contactPostHandler(db));
  router.get("/privacy", privacyHandler(engine));
  router.get("/sitemap.xml", sitemapHandler());

  return router;
}

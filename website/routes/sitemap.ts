// routes/sitemap.ts — Sitemap XML route handler

import type { RouterContext } from "@oak/oak";
import { generateSitemap } from "../content/seo.ts";

const SITE_PAGES = [
  "/",
  "/roadmap",
  "/conops",
  "/docs",
  "/resources",
  "/getting-started",
  "/contributing",
  "/contact",
  "/privacy",
];

export function sitemapHandler() {
  return (ctx: RouterContext<"/sitemap.xml">) => {
    const baseUrl = `${ctx.request.url.protocol}//${ctx.request.url.host}`;
    const xml = generateSitemap(baseUrl, SITE_PAGES);
    ctx.response.body = xml;
    ctx.response.type = "application/xml";
  };
}

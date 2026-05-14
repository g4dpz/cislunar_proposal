// middleware/static.ts — Static file serving middleware

import type { Context, Next } from "@oak/oak";

const MIME_TYPES: Record<string, string> = {
  ".html": "text/html",
  ".css": "text/css",
  ".js": "application/javascript",
  ".json": "application/json",
  ".png": "image/png",
  ".jpg": "image/jpeg",
  ".jpeg": "image/jpeg",
  ".gif": "image/gif",
  ".svg": "image/svg+xml",
  ".ico": "image/x-icon",
  ".woff": "font/woff",
  ".woff2": "font/woff2",
  ".ttf": "font/ttf",
  ".eot": "application/vnd.ms-fontobject",
  ".xml": "application/xml",
  ".txt": "text/plain",
};

/**
 * Serves static files from the public/ directory with correct MIME types
 * and cache headers.
 */
export async function staticFiles(ctx: Context, next: Next): Promise<void> {
  const path = ctx.request.url.pathname;

  // Only serve files that look like static assets (have a file extension)
  const extMatch = path.match(/\.[a-zA-Z0-9]+$/);
  if (!extMatch) {
    await next();
    return;
  }

  // Prevent directory traversal
  if (path.includes("..")) {
    await next();
    return;
  }

  const filePath = `./public${path}`;

  try {
    const file = await Deno.readFile(filePath);
    const ext = extMatch[0].toLowerCase();
    const mimeType = MIME_TYPES[ext] ?? "application/octet-stream";

    ctx.response.body = file;
    ctx.response.type = mimeType;

    // Cache static assets for 1 day in production, no-cache for development
    ctx.response.headers.set(
      "Cache-Control",
      "public, max-age=86400",
    );
  } catch (error) {
    if (error instanceof Deno.errors.NotFound) {
      // File not found — pass to next middleware
      await next();
    } else {
      throw error;
    }
  }
}

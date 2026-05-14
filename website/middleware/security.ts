// middleware/security.ts — Security headers middleware

import type { Context, Next } from "@oak/oak";

/**
 * Adds security-related HTTP headers to all responses.
 */
export async function securityHeaders(ctx: Context, next: Next): Promise<void> {
  await next();

  ctx.response.headers.set("X-Content-Type-Options", "nosniff");
  ctx.response.headers.set("X-Frame-Options", "DENY");
  ctx.response.headers.set(
    "Referrer-Policy",
    "strict-origin-when-cross-origin",
  );
  ctx.response.headers.set(
    "Content-Security-Policy",
    [
      "default-src 'self'",
      "script-src 'self' https://cdn.jsdelivr.net 'unsafe-inline'",
      "style-src 'self' https://cdn.jsdelivr.net 'unsafe-inline'",
      "img-src 'self' data: https://amsat-uk.org https://amsat-dl.org",
      "font-src 'self' https://cdn.jsdelivr.net",
      "connect-src 'self'",
      "frame-ancestors 'none'",
    ].join("; "),
  );
  ctx.response.headers.set(
    "Strict-Transport-Security",
    "max-age=31536000; includeSubDomains",
  );
}

// middleware/logging.ts — Request logging middleware

import type { Context, Next } from "@oak/oak";

/**
 * Logs method, path, status code, and response time for each request.
 */
export async function requestLogger(ctx: Context, next: Next): Promise<void> {
  const start = performance.now();

  await next();

  const duration = (performance.now() - start).toFixed(1);
  const { method } = ctx.request;
  const path = ctx.request.url.pathname;
  const status = ctx.response.status;

  console.log(`${method} ${path} ${status} ${duration}ms`);
}

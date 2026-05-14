// middleware/error.ts — Error handling middleware

import type { Context, Next } from "@oak/oak";

/**
 * Error handling middleware that catches unhandled errors and returns
 * appropriate error pages without leaking internal details.
 */
export async function errorHandler(ctx: Context, next: Next): Promise<void> {
  try {
    await next();

    // Handle 404 for routes that weren't matched
    if (ctx.response.status === 404 && !ctx.response.body) {
      ctx.response.status = 404;
      ctx.response.type = "text/html";
      ctx.response.body = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>404 — Page Not Found</title>
  <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/css/bootstrap.min.css" rel="stylesheet" crossorigin="anonymous">
</head>
<body>
  <main class="container text-center py-5 mt-5">
    <h1 class="display-1">404</h1>
    <p class="lead">Page not found.</p>
    <p>The page you are looking for does not exist or has been moved.</p>
    <a href="/" class="btn btn-primary">Return to Homepage</a>
  </main>
</body>
</html>`;
    }
  } catch (error) {
    console.error("Unhandled server error:", error instanceof Error ? error.stack : error);

    ctx.response.status = 500;
    ctx.response.type = "text/html";
    ctx.response.body = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>500 — Internal Server Error</title>
  <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/css/bootstrap.min.css" rel="stylesheet" crossorigin="anonymous">
</head>
<body>
  <main class="container text-center py-5 mt-5">
    <h1 class="display-1">500</h1>
    <p class="lead">Internal Server Error</p>
    <p>Something went wrong. Please try again later.</p>
    <a href="/" class="btn btn-primary">Return to Homepage</a>
  </main>
</body>
</html>`;
  }
}

// main.ts — Application entry point

import { Application } from "@oak/oak";
import { initHandlebars } from "./views/engine.ts";
import { createRouter } from "./routes/mod.ts";
import { initDatabase } from "./db/mod.ts";
import { requestLogger } from "./middleware/logging.ts";
import { securityHeaders } from "./middleware/security.ts";
import { errorHandler } from "./middleware/error.ts";
import { staticFiles } from "./middleware/static.ts";

// ─── Configuration ────────────────────────────────────────────────────────────

const port = parseInt(Deno.env.get("PORT") ?? "8000", 10);
const dbPath = Deno.env.get("DB_PATH") ?? "./data/contact.db";

// ─── Initialization ──────────────────────────────────────────────────────────

// Ensure data directory exists for SQLite database
try {
  await Deno.mkdir("./data", { recursive: true });
} catch (error) {
  if (!(error instanceof Deno.errors.AlreadyExists)) {
    throw error;
  }
}

// Initialize Handlebars template engine
const engine = await initHandlebars();

// Initialize SQLite database
const db = await initDatabase(dbPath);

// Create router with all routes
const router = createRouter(engine, db);

// ─── Application Setup ───────────────────────────────────────────────────────

const app = new Application();

// Middleware stack (order matters):
// 1. Request logging (outermost — captures timing for all requests)
app.use(requestLogger);

// 2. Security headers (applied to all responses)
app.use(securityHeaders);

// 3. Error handling (catches errors from downstream middleware)
app.use(errorHandler);

// 4. Static file serving (serves CSS, JS, images from public/)
app.use(staticFiles);

// 5. Router (page routes)
app.use(router.routes());
app.use(router.allowedMethods());

// ─── Start Server ────────────────────────────────────────────────────────────

console.log(`Server listening on http://localhost:${port}`);
await app.listen({ port });

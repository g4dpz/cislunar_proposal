// main.ts — Application entry point

import { Application } from "@oak/oak";
import { initHandlebars } from "./views/engine.ts";
import { createRouter } from "./routes/mod.ts";
import { initDatabase } from "./db/mod.ts";
import { requestLogger } from "./middleware/logging.ts";
import { securityHeaders } from "./middleware/security.ts";
import { errorHandler } from "./middleware/error.ts";
import { staticFiles } from "./middleware/static.ts";
import { createAuthMiddleware } from "./middleware/auth.ts";
import { createAuthService } from "./services/auth.ts";
import { createUserService } from "./services/users.ts";
import { createRoleService } from "./services/roles.ts";

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

// Initialize SQLite database (seeds default roles and admin user automatically)
const db = await initDatabase(dbPath);

// Create services
const authService = createAuthService(db);
const userService = createUserService(db);
const roleService = createRoleService(db);

// Create auth middleware (must run before router to populate ctx.state.user)
const authMiddleware = createAuthMiddleware(authService);

// Create router with all routes
const router = createRouter(engine, db, authService, userService, roleService);

// ─── Application Setup ───────────────────────────────────────────────────────

const app = new Application({ proxy: true });

// Middleware stack (order matters):
// 1. Request logging (outermost — captures timing for all requests)
app.use(requestLogger);

// 2. Security headers (applied to all responses)
app.use(securityHeaders);

// 3. Error handling (catches errors from downstream middleware)
app.use(errorHandler);

// 4. Auth middleware (validates session, populates ctx.state.user for all routes)
app.use(authMiddleware);

// 5. Router (page routes — auth-aware, can access ctx.state.user)
app.use(router.routes());
app.use(router.allowedMethods());

// 6. Static file serving (serves CSS, JS, images from public/)
app.use(staticFiles);

// ─── Start Server ────────────────────────────────────────────────────────────

console.log(`Server listening on http://localhost:${port}`);
await app.listen({ port });

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
import {
  loginGetHandler,
  loginPostHandler,
  registerGetHandler,
  registerPostHandler,
  logoutHandler,
} from "./auth.ts";
import {
  profileGetHandler,
  profileUpdateHandler,
  profilePasswordHandler,
} from "./profile.ts";
import {
  usersListHandler,
  userDetailHandler,
  userCreateFormHandler,
  userCreateHandler,
  userEditFormHandler,
  userUpdateHandler,
  userDeleteHandler,
} from "./admin-users.ts";
import {
  rolesListHandler,
  roleDetailHandler,
  roleCreateFormHandler,
  roleCreateHandler,
  roleEditFormHandler,
  roleUpdateHandler,
  roleDeleteHandler,
} from "./admin-roles.ts";
import { statisticsHandler } from "./admin-statistics.ts";
import { requireAuth, requireAdmin, guestOnly } from "../middleware/auth.ts";
import type { AuthService } from "../services/auth.ts";
import type { UserService } from "../services/users.ts";
import type { RoleService } from "../services/roles.ts";

/**
 * Creates and configures the Oak router with all page routes.
 */
export function createRouter(
  engine: HandlebarsEngine,
  db: Database,
  authService: AuthService,
  userService: UserService,
  roleService: RoleService,
): Router {
  const router = new Router();

  // ─── Public Page Routes ───────────────────────────────────────────────────

  router.get("/", homeHandler(engine));
  router.get("/roadmap", roadmapHandler(engine));
  router.get("/conops", conopsHandler(engine));
  router.get("/contact", contactGetHandler(engine));
  router.post("/contact", contactPostHandler(db));
  router.get("/privacy", privacyHandler(engine));
  router.get("/sitemap.xml", sitemapHandler());

  // ─── Auth Routes (guest only for login/register) ──────────────────────────

  router.get("/login", guestOnly(), loginGetHandler(engine));
  router.post("/login", guestOnly(), loginPostHandler(authService, engine));
  router.get("/register", guestOnly(), registerGetHandler(engine));
  router.post("/register", guestOnly(), registerPostHandler(authService, engine));
  router.get("/logout", logoutHandler(authService));

  // ─── Authenticated Routes ─────────────────────────────────────────────────

  router.get("/profile", requireAuth(), profileGetHandler(engine));
  router.post("/profile", requireAuth(), profileUpdateHandler(userService, engine));
  router.post("/profile/password", requireAuth(), profilePasswordHandler(authService, engine));

  // Documentation (requires authentication)
  router.get("/docs", requireAuth(), docsHandler(engine));
  router.get("/docs/:phase/requirements", requireAuth(), docsPhaseHandler(engine));

  // ─── Admin Routes (requires admin role) ───────────────────────────────────

  // Users management
  router.get("/admin/users", requireAdmin(), usersListHandler(userService, engine));
  router.get("/admin/users/new", requireAdmin(), userCreateFormHandler(roleService, engine));
  router.post("/admin/users", requireAdmin(), userCreateHandler(userService, roleService, engine));
  router.get("/admin/users/:id", requireAdmin(), userDetailHandler(userService, engine));
  router.get("/admin/users/:id/edit", requireAdmin(), userEditFormHandler(userService, roleService, engine));
  router.post("/admin/users/:id", requireAdmin(), userUpdateHandler(userService, roleService, engine));
  router.post("/admin/users/:id/delete", requireAdmin(), userDeleteHandler(userService, engine));

  // Roles management
  router.get("/admin/roles", requireAdmin(), rolesListHandler(roleService, engine));
  router.get("/admin/roles/new", requireAdmin(), roleCreateFormHandler(engine));
  router.post("/admin/roles", requireAdmin(), roleCreateHandler(roleService, engine));
  router.get("/admin/roles/:id", requireAdmin(), roleDetailHandler(roleService, engine));
  router.get("/admin/roles/:id/edit", requireAdmin(), roleEditFormHandler(roleService, engine));
  router.post("/admin/roles/:id", requireAdmin(), roleUpdateHandler(roleService, engine));
  router.post("/admin/roles/:id/delete", requireAdmin(), roleDeleteHandler(roleService, engine));

  // Statistics
  router.get("/admin/statistics", requireAdmin(), statisticsHandler(engine));

  return router;
}

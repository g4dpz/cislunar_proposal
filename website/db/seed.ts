// db/seed.ts — Default roles and admin user seeding
// Ensures "admin" and "users" roles exist, and a default admin user is created.
// Idempotent: running multiple times produces the same state.

import { Database } from "@db/sqlite";
import { hash } from "bcrypt";

const DEFAULT_ROLES = [
  { name: "admin", description: "Full system administration access" },
  { name: "users", description: "Standard registered user access" },
] as const;

const DEFAULT_ADMIN = {
  name: "Admin",
  email: "admin@arthur.radio",
  password: "admin123!",
} as const;

/**
 * Seed the database with default roles and admin user.
 * This function is idempotent — it checks for existing records before inserting.
 */
export async function seedDatabase(db: Database): Promise<void> {
  // ─── Create default roles if they don't exist ─────────────────────────────
  for (const role of DEFAULT_ROLES) {
    const existing = db.prepare(
      "SELECT id FROM roles WHERE name = ?",
    ).get(role.name) as { id: number } | undefined;

    if (!existing) {
      db.exec(
        "INSERT INTO roles (name, description) VALUES (?, ?)",
        [role.name, role.description],
      );
    }
  }

  // ─── Create admin user if not exists ────────────────────────────────────────
  const existingUser = db.prepare(
    "SELECT id FROM users WHERE email = ?",
  ).get(DEFAULT_ADMIN.email) as { id: number } | undefined;

  let adminUserId: number;

  if (!existingUser) {
    const passwordHash = await hash(DEFAULT_ADMIN.password);
    db.exec(
      "INSERT INTO users (name, email, password_hash) VALUES (?, ?, ?)",
      [DEFAULT_ADMIN.name, DEFAULT_ADMIN.email, passwordHash],
    );
    adminUserId = db.lastInsertRowId;
  } else {
    adminUserId = existingUser.id;
  }

  // ─── Assign "admin" role to admin user if not already assigned ──────────────
  const adminRole = db.prepare(
    "SELECT id FROM roles WHERE name = ?",
  ).get("admin") as { id: number } | undefined;

  if (adminRole) {
    const existingAssignment = db.prepare(
      "SELECT user_id FROM user_roles WHERE user_id = ? AND role_id = ?",
    ).get(adminUserId, adminRole.id) as { user_id: number } | undefined;

    if (!existingAssignment) {
      db.exec(
        "INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)",
        [adminUserId, adminRole.id],
      );
    }
  }
}

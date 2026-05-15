// services/users.ts — User CRUD operations service
// Handles listing, creating, updating, and deleting users with role assignments.

import { Database } from "@db/sqlite";
import { hash } from "bcrypt";

import type { UserRow, RoleRow } from "../db/mod.ts";

// ─── Types ────────────────────────────────────────────────────────────────────

export interface UserWithRoles {
  id: number;
  name: string;
  email: string;
  roles: RoleRow[];
  createdAt: string;
  updatedAt: string;
}

export interface UserService {
  listUsers(): UserWithRoles[];
  getUser(id: number): UserWithRoles | null;
  createUser(data: {
    name: string;
    email: string;
    password: string;
    roleIds: number[];
  }): Promise<UserRow>;
  updateUser(
    id: number,
    data: { name: string; email: string; roleIds: number[] },
  ): UserRow;
  deleteUser(id: number): void;
  isEmailAvailable(email: string, excludeUserId?: number): boolean;
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

/**
 * Fetch roles for a given user ID.
 */
function getRolesForUser(db: Database, userId: number): RoleRow[] {
  const rows = db.prepare(
    `SELECT r.id, r.name, r.description
     FROM roles r
     INNER JOIN user_roles ur ON ur.role_id = r.id
     WHERE ur.user_id = ?`,
  ).all(userId) as Array<{ id: number; name: string; description: string }>;

  return rows.map((r) => ({
    id: r.id,
    name: r.name,
    description: r.description,
  }));
}

// ─── Implementation ───────────────────────────────────────────────────────────

/**
 * Create a UserService instance backed by the given SQLite database.
 */
export function createUserService(db: Database): UserService {
  return {
    listUsers(): UserWithRoles[] {
      const rows = db.prepare(
        `SELECT id, name, email, created_at, updated_at
         FROM users
         ORDER BY created_at DESC`,
      ).all() as Array<{
        id: number;
        name: string;
        email: string;
        created_at: string;
        updated_at: string;
      }>;

      return rows.map((row) => ({
        id: row.id,
        name: row.name,
        email: row.email,
        roles: getRolesForUser(db, row.id),
        createdAt: row.created_at,
        updatedAt: row.updated_at,
      }));
    },

    getUser(id: number): UserWithRoles | null {
      const row = db.prepare(
        `SELECT id, name, email, created_at, updated_at
         FROM users
         WHERE id = ?`,
      ).get(id) as {
        id: number;
        name: string;
        email: string;
        created_at: string;
        updated_at: string;
      } | undefined;

      if (!row) {
        return null;
      }

      return {
        id: row.id,
        name: row.name,
        email: row.email,
        roles: getRolesForUser(db, row.id),
        createdAt: row.created_at,
        updatedAt: row.updated_at,
      };
    },

    async createUser(data: {
      name: string;
      email: string;
      password: string;
      roleIds: number[];
    }): Promise<UserRow> {
      // Hash the password with bcrypt
      const passwordHash = await hash(data.password);

      // Insert the user record
      db.exec(
        "INSERT INTO users (name, email, password_hash) VALUES (?, ?, ?)",
        [data.name, data.email, passwordHash],
      );
      const userId = db.lastInsertRowId;

      // Assign roles
      for (const roleId of data.roleIds) {
        db.exec(
          "INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)",
          [userId, roleId],
        );
      }

      // Retrieve and return the created user
      const user = db.prepare(
        `SELECT id, name, email, password_hash, created_at, updated_at
         FROM users WHERE id = ?`,
      ).get(userId) as {
        id: number;
        name: string;
        email: string;
        password_hash: string;
        created_at: string;
        updated_at: string;
      };

      return {
        id: user.id,
        name: user.name,
        email: user.email,
        passwordHash: user.password_hash,
        createdAt: user.created_at,
        updatedAt: user.updated_at,
      };
    },

    updateUser(
      id: number,
      data: { name: string; email: string; roleIds: number[] },
    ): UserRow {
      // Update user fields and set updated_at timestamp
      db.exec(
        `UPDATE users SET name = ?, email = ?, updated_at = datetime('now') WHERE id = ?`,
        [data.name, data.email, id],
      );

      // Replace role assignments: remove existing, then insert new
      db.exec("DELETE FROM user_roles WHERE user_id = ?", [id]);
      for (const roleId of data.roleIds) {
        db.exec(
          "INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)",
          [id, roleId],
        );
      }

      // Retrieve and return the updated user
      const user = db.prepare(
        `SELECT id, name, email, password_hash, created_at, updated_at
         FROM users WHERE id = ?`,
      ).get(id) as {
        id: number;
        name: string;
        email: string;
        password_hash: string;
        created_at: string;
        updated_at: string;
      };

      return {
        id: user.id,
        name: user.name,
        email: user.email,
        passwordHash: user.password_hash,
        createdAt: user.created_at,
        updatedAt: user.updated_at,
      };
    },

    deleteUser(id: number): void {
      // CASCADE delete handles user_roles cleanup automatically
      db.exec("DELETE FROM users WHERE id = ?", [id]);
    },

    isEmailAvailable(email: string, excludeUserId?: number): boolean {
      if (excludeUserId !== undefined) {
        const row = db.prepare(
          "SELECT id FROM users WHERE email = ? AND id != ?",
        ).get(email, excludeUserId) as { id: number } | undefined;
        return row === undefined;
      }

      const row = db.prepare(
        "SELECT id FROM users WHERE email = ?",
      ).get(email) as { id: number } | undefined;
      return row === undefined;
    },
  };
}

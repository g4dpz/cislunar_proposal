// services/roles.ts — Role CRUD operations service
// Handles listing, creating, updating, and deleting roles with user assignment checks.

import { Database } from "@db/sqlite";

import type { RoleRow } from "../db/mod.ts";

// ─── Types ────────────────────────────────────────────────────────────────────

export interface RoleWithCount {
  id: number;
  name: string;
  description: string;
  userCount: number;
}

export interface RoleWithUsers {
  id: number;
  name: string;
  description: string;
  users: Array<{ id: number; name: string; email: string }>;
}

export interface RoleService {
  listRoles(): RoleWithCount[];
  getRole(id: number): RoleWithUsers | null;
  createRole(data: { name: string; description: string }): RoleRow;
  updateRole(id: number, data: { name: string; description: string }): RoleRow;
  deleteRole(id: number): { success: boolean; error?: string };
  isNameAvailable(name: string, excludeRoleId?: number): boolean;
}

// ─── Implementation ───────────────────────────────────────────────────────────

/**
 * Create a RoleService instance backed by the given SQLite database.
 */
export function createRoleService(db: Database): RoleService {
  return {
    listRoles(): RoleWithCount[] {
      const rows = db.prepare(
        `SELECT r.id, r.name, r.description,
                COUNT(ur.user_id) AS user_count
         FROM roles r
         LEFT JOIN user_roles ur ON ur.role_id = r.id
         GROUP BY r.id
         ORDER BY r.name ASC`,
      ).all() as Array<{
        id: number;
        name: string;
        description: string;
        user_count: number;
      }>;

      return rows.map((row) => ({
        id: row.id,
        name: row.name,
        description: row.description,
        userCount: row.user_count,
      }));
    },

    getRole(id: number): RoleWithUsers | null {
      const row = db.prepare(
        `SELECT id, name, description FROM roles WHERE id = ?`,
      ).get(id) as { id: number; name: string; description: string } | undefined;

      if (!row) {
        return null;
      }

      const users = db.prepare(
        `SELECT u.id, u.name, u.email
         FROM users u
         INNER JOIN user_roles ur ON ur.user_id = u.id
         WHERE ur.role_id = ?`,
      ).all(id) as Array<{ id: number; name: string; email: string }>;

      return {
        id: row.id,
        name: row.name,
        description: row.description,
        users: users.map((u) => ({
          id: u.id,
          name: u.name,
          email: u.email,
        })),
      };
    },

    createRole(data: { name: string; description: string }): RoleRow {
      db.exec(
        "INSERT INTO roles (name, description) VALUES (?, ?)",
        [data.name, data.description],
      );
      const roleId = db.lastInsertRowId;

      const role = db.prepare(
        `SELECT id, name, description FROM roles WHERE id = ?`,
      ).get(roleId) as { id: number; name: string; description: string };

      return {
        id: role.id,
        name: role.name,
        description: role.description,
      };
    },

    updateRole(
      id: number,
      data: { name: string; description: string },
    ): RoleRow {
      db.exec(
        `UPDATE roles SET name = ?, description = ? WHERE id = ?`,
        [data.name, data.description, id],
      );

      const role = db.prepare(
        `SELECT id, name, description FROM roles WHERE id = ?`,
      ).get(id) as { id: number; name: string; description: string };

      return {
        id: role.id,
        name: role.name,
        description: role.description,
      };
    },

    deleteRole(id: number): { success: boolean; error?: string } {
      // Check if any users are assigned to this role
      const assignment = db.prepare(
        `SELECT COUNT(*) AS count FROM user_roles WHERE role_id = ?`,
      ).get(id) as { count: number };

      if (assignment.count > 0) {
        return {
          success: false,
          error: "Cannot delete role with assigned users",
        };
      }

      db.exec("DELETE FROM roles WHERE id = ?", [id]);
      return { success: true };
    },

    isNameAvailable(name: string, excludeRoleId?: number): boolean {
      if (excludeRoleId !== undefined) {
        const row = db.prepare(
          "SELECT id FROM roles WHERE name = ? AND id != ?",
        ).get(name, excludeRoleId) as { id: number } | undefined;
        return row === undefined;
      }

      const row = db.prepare(
        "SELECT id FROM roles WHERE name = ?",
      ).get(name) as { id: number } | undefined;
      return row === undefined;
    },
  };
}

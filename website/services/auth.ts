// services/auth.ts — Authentication and session management service
// Handles registration, login, logout, session validation, and password operations.

import { Database } from "@db/sqlite";
import { hash, compare } from "bcrypt";

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

export interface AuthResult {
  sessionToken: string;
  user: UserRow;
}

export interface AuthService {
  register(
    name: string,
    email: string,
    password: string,
  ): Promise<AuthResult>;
  login(
    email: string,
    password: string,
  ): Promise<AuthResult | null>;
  logout(sessionToken: string): Promise<void>;
  validateSession(sessionToken: string): Promise<UserWithRoles | null>;
  changePassword(
    userId: number,
    currentPassword: string,
    newPassword: string,
  ): Promise<boolean>;
  cleanExpiredSessions(): Promise<void>;
}

// ─── Implementation ───────────────────────────────────────────────────────────

/**
 * Create an AuthService instance backed by the given SQLite database.
 */
export function createAuthService(db: Database): AuthService {
  return {
    async register(
      name: string,
      email: string,
      password: string,
    ): Promise<AuthResult> {
      // Validate password minimum length
      if (password.length < 8) {
        throw new Error("Password must be at least 8 characters");
      }

      // Check email uniqueness
      const existing = db.prepare(
        "SELECT id FROM users WHERE email = ?",
      ).get(email) as { id: number } | undefined;

      if (existing) {
        throw new Error("Email is already registered");
      }

      // Hash password and create user
      const passwordHash = await hash(password);
      db.exec(
        "INSERT INTO users (name, email, password_hash) VALUES (?, ?, ?)",
        [name, email, passwordHash],
      );
      const userId = db.lastInsertRowId;

      // Assign "users" role
      const usersRole = db.prepare(
        "SELECT id FROM roles WHERE name = ?",
      ).get("users") as { id: number } | undefined;

      if (usersRole) {
        db.exec(
          "INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)",
          [userId, usersRole.id],
        );
      }

      // Create session
      const sessionToken = crypto.randomUUID();
      const expiresAt = new Date(Date.now() + 60 * 60 * 1000).toISOString();
      db.exec(
        "INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)",
        [sessionToken, userId, expiresAt],
      );

      // Retrieve created user
      const user = db.prepare(
        "SELECT id, name, email, password_hash, created_at, updated_at FROM users WHERE id = ?",
      ).get(userId) as {
        id: number;
        name: string;
        email: string;
        password_hash: string;
        created_at: string;
        updated_at: string;
      };

      return {
        sessionToken,
        user: {
          id: user.id,
          name: user.name,
          email: user.email,
          passwordHash: user.password_hash,
          createdAt: user.created_at,
          updatedAt: user.updated_at,
        },
      };
    },

    async login(
      email: string,
      password: string,
    ): Promise<AuthResult | null> {
      // Find user by email
      const row = db.prepare(
        "SELECT id, name, email, password_hash, created_at, updated_at FROM users WHERE email = ?",
      ).get(email) as {
        id: number;
        name: string;
        email: string;
        password_hash: string;
        created_at: string;
        updated_at: string;
      } | undefined;

      if (!row) {
        return null;
      }

      // Verify password against stored hash
      const valid = await compare(password, row.password_hash);
      if (!valid) {
        return null;
      }

      // Create session with 1-hour expiry
      const sessionToken = crypto.randomUUID();
      const expiresAt = new Date(Date.now() + 60 * 60 * 1000).toISOString();
      db.exec(
        "INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)",
        [sessionToken, row.id, expiresAt],
      );

      return {
        sessionToken,
        user: {
          id: row.id,
          name: row.name,
          email: row.email,
          passwordHash: row.password_hash,
          createdAt: row.created_at,
          updatedAt: row.updated_at,
        },
      };
    },

    async logout(sessionToken: string): Promise<void> {
      db.exec("DELETE FROM sessions WHERE id = ?", [sessionToken]);
    },

    async validateSession(
      sessionToken: string,
    ): Promise<UserWithRoles | null> {
      // Check session exists and is not expired
      const session = db.prepare(
        "SELECT user_id, expires_at FROM sessions WHERE id = ?",
      ).get(sessionToken) as {
        user_id: number;
        expires_at: string;
      } | undefined;

      if (!session) {
        return null;
      }

      // Check expiry
      const now = new Date();
      const expiresAt = new Date(session.expires_at);
      if (now >= expiresAt) {
        // Session expired — clean it up
        db.exec("DELETE FROM sessions WHERE id = ?", [sessionToken]);
        return null;
      }

      // Get user
      const user = db.prepare(
        "SELECT id, name, email, created_at, updated_at FROM users WHERE id = ?",
      ).get(session.user_id) as {
        id: number;
        name: string;
        email: string;
        created_at: string;
        updated_at: string;
      } | undefined;

      if (!user) {
        return null;
      }

      // Get user's roles (join user_roles and roles tables)
      const roleRows = db.prepare(
        `SELECT r.id, r.name, r.description
         FROM roles r
         INNER JOIN user_roles ur ON ur.role_id = r.id
         WHERE ur.user_id = ?`,
      ).all(user.id) as Array<{
        id: number;
        name: string;
        description: string;
      }>;

      const roles: RoleRow[] = roleRows.map((r) => ({
        id: r.id,
        name: r.name,
        description: r.description,
      }));

      return {
        id: user.id,
        name: user.name,
        email: user.email,
        roles,
        createdAt: user.created_at,
        updatedAt: user.updated_at,
      };
    },

    async changePassword(
      userId: number,
      currentPassword: string,
      newPassword: string,
    ): Promise<boolean> {
      // Validate new password minimum length
      if (newPassword.length < 8) {
        throw new Error("Password must be at least 8 characters");
      }

      // Get current password hash
      const row = db.prepare(
        "SELECT password_hash FROM users WHERE id = ?",
      ).get(userId) as { password_hash: string } | undefined;

      if (!row) {
        return false;
      }

      // Verify current password
      const valid = await compare(currentPassword, row.password_hash);
      if (!valid) {
        return false;
      }

      // Hash new password and update
      const newHash = await hash(newPassword);
      db.exec(
        "UPDATE users SET password_hash = ?, updated_at = datetime('now') WHERE id = ?",
        [newHash, userId],
      );

      return true;
    },

    async cleanExpiredSessions(): Promise<void> {
      const now = new Date().toISOString();
      db.exec("DELETE FROM sessions WHERE expires_at <= ?", [now]);
    },
  };
}

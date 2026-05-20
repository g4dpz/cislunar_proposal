// db/mod.ts — SQLite database module for contact form submissions, users, roles, and sessions

import { Database } from "@db/sqlite";
import { seedDatabase } from "./seed.ts";

export { seedDatabase } from "./seed.ts";

// ─── Interfaces ───────────────────────────────────────────────────────────────

export interface ContactSubmission {
  id: number;
  name: string;
  callsignOrOrg: string;
  areaOfInterest: string;
  message: string;
  submittedAt: string;
}

export type NewContactSubmission = Omit<ContactSubmission, "id" | "submittedAt">;

export interface UserRow {
  id: number;
  name: string;
  email: string;
  passwordHash: string;
  createdAt: string;
  updatedAt: string;
}

export interface RoleRow {
  id: number;
  name: string;
  description: string;
}

export interface UserRoleRow {
  userId: number;
  roleId: number;
}

export interface SessionRow {
  id: string;
  userId: number;
  expiresAt: string;
  createdAt: string;
}

export interface OutreachContactRow {
  id: number;
  name: string;
  contact_type: string;
  status: string;
  contact_url: string | null;
  callsign: string | null;
  email: string | null;
  notes: string | null;
  contacted_by: string | null;
  contacted_date: string | null;
  created_at: string;
  last_updated: string;
}

// ─── Database Initialization ──────────────────────────────────────────────────

/**
 * Initialize the SQLite database and create tables if they don't exist.
 * Seeds default roles and admin user.
 */
export async function initDatabase(dbPath: string): Promise<Database> {
  const db = new Database(dbPath);

  // Enable WAL mode and foreign keys
  db.exec("PRAGMA journal_mode = WAL");
  db.exec("PRAGMA foreign_keys = ON");

  db.exec(`
    CREATE TABLE IF NOT EXISTS contact_submissions (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      name TEXT NOT NULL,
      callsign_or_org TEXT,
      area_of_interest TEXT,
      message TEXT NOT NULL,
      submitted_at TEXT DEFAULT (datetime('now'))
    );
  `);

  db.exec(`
    CREATE TABLE IF NOT EXISTS users (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      name TEXT NOT NULL,
      email TEXT NOT NULL UNIQUE,
      password_hash TEXT NOT NULL,
      created_at TEXT DEFAULT (datetime('now')),
      updated_at TEXT DEFAULT (datetime('now'))
    );
  `);

  db.exec(`
    CREATE TABLE IF NOT EXISTS roles (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      name TEXT NOT NULL UNIQUE,
      description TEXT NOT NULL DEFAULT ''
    );
  `);

  db.exec(`
    CREATE TABLE IF NOT EXISTS user_roles (
      user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      role_id INTEGER NOT NULL REFERENCES roles(id) ON DELETE RESTRICT,
      PRIMARY KEY (user_id, role_id)
    );
  `);

  db.exec(`
    CREATE TABLE IF NOT EXISTS sessions (
      id TEXT PRIMARY KEY,
      user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      expires_at TEXT NOT NULL,
      created_at TEXT DEFAULT (datetime('now'))
    );
  `);

  // Index for session expiry cleanup
  db.exec(`CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);`);

  // Index for user email lookups
  db.exec(`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);`);

  db.exec(`
    CREATE TABLE IF NOT EXISTS outreach_contacts (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      name TEXT NOT NULL,
      contact_type TEXT NOT NULL CHECK(contact_type IN (
        'facebook_group', 'mailing_list', 'club', 'university',
        'individual', 'cubesat_team', 'organisation'
      )),
      status TEXT NOT NULL DEFAULT 'identified' CHECK(status IN (
        'identified', 'contacted', 'responded', 'collaborating'
      )),
      contact_url TEXT,
      callsign TEXT,
      email TEXT,
      notes TEXT,
      contacted_by TEXT,
      contacted_date TEXT,
      created_at TEXT DEFAULT (datetime('now')),
      last_updated TEXT DEFAULT (datetime('now')),
      UNIQUE(name COLLATE NOCASE, contact_type)
    );
  `);

  // Index for status filtering (most common query pattern)
  db.exec(`CREATE INDEX IF NOT EXISTS idx_outreach_status ON outreach_contacts(status);`);

  // Index for type filtering
  db.exec(`CREATE INDEX IF NOT EXISTS idx_outreach_type ON outreach_contacts(contact_type);`);

  // Seed default roles and admin user
  await seedDatabase(db);

  return db;
}

// ─── Queries ──────────────────────────────────────────────────────────────────

/**
 * Save a contact form submission to the database.
 */
export function saveContactSubmission(
  db: Database,
  submission: NewContactSubmission,
): void {
  db.exec(
    `INSERT INTO contact_submissions (name, callsign_or_org, area_of_interest, message)
     VALUES (?, ?, ?, ?)`,
    [
      submission.name,
      submission.callsignOrOrg,
      submission.areaOfInterest,
      submission.message,
    ],
  );
}

/**
 * Retrieve all contact submissions from the database.
 */
export function getContactSubmissions(db: Database): ContactSubmission[] {
  const rows = db.prepare(
    `SELECT id, name, callsign_or_org, area_of_interest, message, submitted_at
     FROM contact_submissions
     ORDER BY submitted_at DESC`,
  ).all() as Array<{
    id: number;
    name: string;
    callsign_or_org: string;
    area_of_interest: string;
    message: string;
    submitted_at: string;
  }>;

  return rows.map((row) => ({
    id: row.id,
    name: row.name,
    callsignOrOrg: row.callsign_or_org ?? "",
    areaOfInterest: row.area_of_interest ?? "",
    message: row.message,
    submittedAt: row.submitted_at,
  }));
}

// db/mod.ts — SQLite database module for contact form submissions

import { Database } from "@db/sqlite";

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

// ─── Database Initialization ──────────────────────────────────────────────────

/**
 * Initialize the SQLite database and create tables if they don't exist.
 */
export function initDatabase(dbPath: string): Database {
  const db = new Database(dbPath);

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

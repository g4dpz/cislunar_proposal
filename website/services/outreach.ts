// services/outreach.ts — Outreach contact CRUD operations service
// Handles listing, creating, updating, and deleting outreach contacts,
// status pipeline advancement, duplicate detection, and statistics.

import { Database } from "@db/sqlite";

import type { OutreachContactRow } from "../db/mod.ts";

// ─── Types ────────────────────────────────────────────────────────────────────

export type ContactType =
  | "facebook_group"
  | "mailing_list"
  | "club"
  | "university"
  | "individual"
  | "cubesat_team"
  | "organisation";

export type OutreachStatus =
  | "identified"
  | "contacted"
  | "responded"
  | "collaborating";

export interface ContactRecord {
  id: number;
  name: string;
  contactType: ContactType;
  status: OutreachStatus;
  contactUrl: string | null;
  callsign: string | null;
  email: string | null;
  notes: string | null;
  contactedBy: string | null;
  contactedDate: string | null;
  createdAt: string;
  lastUpdated: string;
}

export interface CreateContactData {
  name: string;
  contactType: ContactType;
  status?: OutreachStatus;
  contactUrl?: string;
  callsign?: string;
  email?: string;
  notes?: string;
  contactedBy?: string;
  contactedDate?: string;
}

export interface UpdateContactData {
  name: string;
  contactType: ContactType;
  status: OutreachStatus;
  contactUrl?: string;
  callsign?: string;
  email?: string;
  notes?: string;
  contactedBy?: string;
  contactedDate?: string;
}

export interface OutreachStats {
  total: number;
  byStatus: Record<OutreachStatus, number>;
  byType: Record<ContactType, number>;
}

export interface OutreachService {
  listContacts(filters?: {
    status?: OutreachStatus;
    contactType?: ContactType;
  }): ContactRecord[];
  getContact(id: number): ContactRecord | null;
  createContact(data: CreateContactData): ContactRecord;
  updateContact(id: number, data: UpdateContactData): ContactRecord;
  deleteContact(id: number): void;
  advanceStatus(id: number): ContactRecord;
  isDuplicate(
    name: string,
    contactType: ContactType,
    excludeId?: number,
  ): boolean;
  searchSimilar(name: string): ContactRecord[];
  getStats(): OutreachStats;
  getCollaborators(): ContactRecord[];
}


// ─── Constants ────────────────────────────────────────────────────────────────

const STATUS_PIPELINE: OutreachStatus[] = [
  "identified",
  "contacted",
  "responded",
  "collaborating",
];

// ─── Helpers ──────────────────────────────────────────────────────────────────

/**
 * Map a raw database row to a ContactRecord with camelCase fields.
 */
function rowToContact(row: OutreachContactRow): ContactRecord {
  return {
    id: row.id,
    name: row.name,
    contactType: row.contact_type as ContactType,
    status: row.status as OutreachStatus,
    contactUrl: row.contact_url,
    callsign: row.callsign,
    email: row.email,
    notes: row.notes,
    contactedBy: row.contacted_by,
    contactedDate: row.contacted_date,
    createdAt: row.created_at,
    lastUpdated: row.last_updated,
  };
}

// ─── Implementation ───────────────────────────────────────────────────────────

/**
 * Create an OutreachService instance backed by the given SQLite database.
 */
export function createOutreachService(db: Database): OutreachService {
  return {
    listContacts(filters?: {
      status?: OutreachStatus;
      contactType?: ContactType;
    }): ContactRecord[] {
      let sql =
        `SELECT id, name, contact_type, status, contact_url, callsign, email, notes, contacted_by, contacted_date, created_at, last_updated FROM outreach_contacts`;
      const conditions: string[] = [];
      const params: (string | number | null)[] = [];

      if (filters?.status) {
        conditions.push("status = ?");
        params.push(filters.status);
      }
      if (filters?.contactType) {
        conditions.push("contact_type = ?");
        params.push(filters.contactType);
      }

      if (conditions.length > 0) {
        sql += " WHERE " + conditions.join(" AND ");
      }

      sql += " ORDER BY last_updated DESC";

      const rows = db.prepare(sql).all(...params) as OutreachContactRow[];
      return rows.map(rowToContact);
    },

    getContact(id: number): ContactRecord | null {
      const row = db.prepare(
        `SELECT id, name, contact_type, status, contact_url, callsign, email, notes, contacted_by, contacted_date, created_at, last_updated
         FROM outreach_contacts WHERE id = ?`,
      ).get(id) as OutreachContactRow | undefined;

      if (!row) {
        return null;
      }

      return rowToContact(row);
    },

    createContact(data: CreateContactData): ContactRecord {
      // Enforce name+type uniqueness (case-insensitive)
      if (this.isDuplicate(data.name, data.contactType)) {
        throw new Error(
          `A contact with name "${data.name}" and type "${data.contactType}" already exists`,
        );
      }

      const status = data.status ?? "identified";

      db.exec(
        `INSERT INTO outreach_contacts (name, contact_type, status, contact_url, callsign, email, notes, contacted_by, contacted_date)
         VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
        [
          data.name,
          data.contactType,
          status,
          data.contactUrl ?? null,
          data.callsign ?? null,
          data.email ?? null,
          data.notes ?? null,
          data.contactedBy ?? null,
          data.contactedDate ?? null,
        ],
      );

      const id = db.lastInsertRowId;
      return this.getContact(Number(id))!;
    },

    updateContact(id: number, data: UpdateContactData): ContactRecord {
      // Verify the contact exists
      const existing = this.getContact(id);
      if (!existing) {
        throw new Error(`Contact with id ${id} not found`);
      }

      // Enforce name+type uniqueness (case-insensitive), excluding current record
      if (this.isDuplicate(data.name, data.contactType, id)) {
        throw new Error(
          `A contact with name "${data.name}" and type "${data.contactType}" already exists`,
        );
      }

      db.exec(
        `UPDATE outreach_contacts
         SET name = ?, contact_type = ?, status = ?, contact_url = ?, callsign = ?, email = ?, notes = ?, contacted_by = ?, contacted_date = ?, last_updated = datetime('now')
         WHERE id = ?`,
        [
          data.name,
          data.contactType,
          data.status,
          data.contactUrl ?? null,
          data.callsign ?? null,
          data.email ?? null,
          data.notes ?? null,
          data.contactedBy ?? null,
          data.contactedDate ?? null,
          id,
        ],
      );

      return this.getContact(id)!;
    },

    deleteContact(id: number): void {
      db.exec("DELETE FROM outreach_contacts WHERE id = ?", [id]);
    },

    advanceStatus(id: number): ContactRecord {
      const contact = this.getContact(id);
      if (!contact) {
        throw new Error(`Contact with id ${id} not found`);
      }

      const currentIndex = STATUS_PIPELINE.indexOf(contact.status);
      if (currentIndex === STATUS_PIPELINE.length - 1) {
        throw new Error(
          `Cannot advance status: contact is already at "${contact.status}"`,
        );
      }

      const nextStatus = STATUS_PIPELINE[currentIndex + 1];

      db.exec(
        `UPDATE outreach_contacts SET status = ?, last_updated = datetime('now') WHERE id = ?`,
        [nextStatus, id],
      );

      return this.getContact(id)!;
    },

    isDuplicate(
      name: string,
      contactType: ContactType,
      excludeId?: number,
    ): boolean {
      if (excludeId !== undefined) {
        const row = db.prepare(
          `SELECT id FROM outreach_contacts WHERE name = ? COLLATE NOCASE AND contact_type = ? AND id != ?`,
        ).get(name, contactType, excludeId) as { id: number } | undefined;
        return row !== undefined;
      }

      const row = db.prepare(
        `SELECT id FROM outreach_contacts WHERE name = ? COLLATE NOCASE AND contact_type = ?`,
      ).get(name, contactType) as { id: number } | undefined;
      return row !== undefined;
    },

    searchSimilar(name: string): ContactRecord[] {
      const rows = db.prepare(
        `SELECT id, name, contact_type, status, contact_url, callsign, email, notes, contacted_by, contacted_date, created_at, last_updated
         FROM outreach_contacts
         WHERE name LIKE ?
         ORDER BY name ASC`,
      ).all(`%${name}%`) as OutreachContactRow[];

      return rows.map(rowToContact);
    },

    getStats(): OutreachStats {
      // Count by status
      const statusRows = db.prepare(
        `SELECT status, COUNT(*) AS count FROM outreach_contacts GROUP BY status`,
      ).all() as Array<{ status: string; count: number }>;

      const byStatus: Record<OutreachStatus, number> = {
        identified: 0,
        contacted: 0,
        responded: 0,
        collaborating: 0,
      };
      for (const row of statusRows) {
        byStatus[row.status as OutreachStatus] = row.count;
      }

      // Count by type
      const typeRows = db.prepare(
        `SELECT contact_type, COUNT(*) AS count FROM outreach_contacts GROUP BY contact_type`,
      ).all() as Array<{ contact_type: string; count: number }>;

      const byType: Record<ContactType, number> = {
        facebook_group: 0,
        mailing_list: 0,
        club: 0,
        university: 0,
        individual: 0,
        cubesat_team: 0,
        organisation: 0,
      };
      for (const row of typeRows) {
        byType[row.contact_type as ContactType] = row.count;
      }

      // Total count
      const totalRow = db.prepare(
        `SELECT COUNT(*) AS count FROM outreach_contacts`,
      ).get() as { count: number };

      return {
        total: totalRow.count,
        byStatus,
        byType,
      };
    },

    getCollaborators(): ContactRecord[] {
      const rows = db.prepare(
        `SELECT id, name, contact_type, status, contact_url, callsign, email, notes, contacted_by, contacted_date, created_at, last_updated
         FROM outreach_contacts
         WHERE status = 'collaborating'
         ORDER BY last_updated DESC`,
      ).all() as OutreachContactRow[];

      return rows.map(rowToContact);
    },
  };
}

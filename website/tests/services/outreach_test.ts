/**
 * Property-Based Tests for Outreach Service
 *
 * Property 1: Contact Record CRUD Round-Trip
 * Property 2: Name+Type Uniqueness Invariant (Case-Insensitive)
 * Property 3: Status Pipeline Advancement
 * Property 4: Default Status Invariant
 * Property 5: Contact List Ordering
 * Property 6: Filter Correctness
 * Property 7: Public View Shows Only Collaborators
 * Property 8: Statistics Consistency
 *
 * Feature: outreach-tracker
 */

import {
  assertEquals,
} from "https://deno.land/std@0.224.0/assert/mod.ts";
import fc from "fast-check";
import { Database } from "@db/sqlite";
import {
  type ContactType,
  createOutreachService,
  type OutreachStatus,
} from "../../services/outreach.ts";

// ─── Helpers ──────────────────────────────────────────────────────────────────

const CONTACT_TYPES: ContactType[] = [
  "facebook_group",
  "mailing_list",
  "club",
  "university",
  "individual",
  "cubesat_team",
  "organisation",
];

const STATUSES: OutreachStatus[] = [
  "identified",
  "contacted",
  "responded",
  "collaborating",
];

function createTestDb(): Database {
  const db = new Database(":memory:");
  db.exec("PRAGMA journal_mode = WAL");
  db.exec("PRAGMA foreign_keys = ON");
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
  db.exec(`CREATE INDEX IF NOT EXISTS idx_outreach_status ON outreach_contacts(status);`);
  db.exec(`CREATE INDEX IF NOT EXISTS idx_outreach_type ON outreach_contacts(contact_type);`);
  return db;
}

/** Arbitrary for generating a valid contact name. */
const arbContactName: fc.Arbitrary<string> = fc
  .stringMatching(/^[a-zA-Z][a-zA-Z0-9 ]{0,30}$/)
  .filter((s: string) => s.trim().length > 0)
  .map((s: string) => s.trim());

/** Arbitrary for generating a ContactType. */
const arbContactType: fc.Arbitrary<ContactType> = fc.constantFrom(...CONTACT_TYPES);

/** Arbitrary for generating an OutreachStatus. */
const arbStatus: fc.Arbitrary<OutreachStatus> = fc.constantFrom(...STATUSES);

// ─── Property 1: Contact Record CRUD Round-Trip ───────────────────────────────
/**
 * Property 1: Contact Record CRUD Round-Trip
 *
 * Creating a contact with data D then reading it returns D (with generated id
 * and timestamps). Updating a contact then reading it reflects the changes.
 * Deleting a contact then reading it returns null.
 *
 * **Validates: Requirements 1.1, 2.1, 4.1, 5.2**
 */
Deno.test("Property 1: CRUD Round-Trip - create/read/update/delete cycle preserves data", async () => {
  await fc.assert(
    fc.asyncProperty(
      arbContactName,
      arbContactType,
      arbContactName,
      arbContactType,
      arbStatus,
      async (
        name: string,
        contactType: ContactType,
        updatedName: string,
        updatedType: ContactType,
        updatedStatus: OutreachStatus,
      ) => {
        const db = createTestDb();
        const service = createOutreachService(db);

        // CREATE then READ
        const created = service.createContact({ name, contactType });
        assertEquals(created.name, name);
        assertEquals(created.contactType, contactType);
        assertEquals(created.status, "identified"); // default status
        assertEquals(created.contactUrl, null);
        assertEquals(created.callsign, null);
        assertEquals(created.email, null);
        assertEquals(created.notes, null);
        assertEquals(created.contactedBy, null);

        const read = service.getContact(created.id);
        assertEquals(read !== null, true);
        assertEquals(read!.id, created.id);
        assertEquals(read!.name, created.name);
        assertEquals(read!.contactType, created.contactType);

        // UPDATE then READ
        const updateData = {
          name: updatedName,
          contactType: updatedType,
          status: updatedStatus,
        };

        const updated = service.updateContact(created.id, updateData);
        assertEquals(updated.name, updatedName);
        assertEquals(updated.contactType, updatedType);
        assertEquals(updated.status, updatedStatus);

        const readAfterUpdate = service.getContact(created.id);
        assertEquals(readAfterUpdate!.name, updatedName);
        assertEquals(readAfterUpdate!.contactType, updatedType);
        assertEquals(readAfterUpdate!.status, updatedStatus);

        // DELETE then READ
        service.deleteContact(created.id);
        const readAfterDelete = service.getContact(created.id);
        assertEquals(readAfterDelete, null);

        db.close();
      },
    ),
    { numRuns: 30 },
  );
});

// ─── Property 2: Name+Type Uniqueness Invariant (Case-Insensitive) ────────────
/**
 * Property 2: Name+Type Uniqueness Invariant (Case-Insensitive)
 *
 * No two records share the same name (case-insensitive) and contact_type.
 * Attempting to create or update a record to violate this always fails.
 *
 * **Validates: Requirements 1.4, 2.2, 4.4, 7.2, 7.3**
 */
Deno.test("Property 2: Name+Type Uniqueness - case-insensitive duplicate detection", async () => {
  await fc.assert(
    fc.asyncProperty(
      arbContactName,
      arbContactType,
      fc.constantFrom("upper", "lower", "mixed"),
      async (name: string, contactType: ContactType, caseVariant: string) => {
        const db = createTestDb();
        const service = createOutreachService(db);

        // Create the first contact
        service.createContact({ name, contactType });

        // Generate a case variant of the name
        let variantName: string;
        if (caseVariant === "upper") {
          variantName = name.toUpperCase();
        } else if (caseVariant === "lower") {
          variantName = name.toLowerCase();
        } else {
          variantName = name
            .split("")
            .map((c: string, i: number) =>
              i % 2 === 0 ? c.toUpperCase() : c.toLowerCase()
            )
            .join("");
        }

        // Attempting to create a duplicate (same name case-insensitive + same type) should throw
        let threw = false;
        try {
          service.createContact({ name: variantName, contactType });
        } catch (_e) {
          threw = true;
        }
        assertEquals(threw, true, "Creating a case-insensitive duplicate should throw");

        // isDuplicate should detect it
        assertEquals(
          service.isDuplicate(variantName, contactType),
          true,
          "isDuplicate should detect case-insensitive match",
        );

        // Creating with a DIFFERENT type should succeed
        const otherType = CONTACT_TYPES.find((t) => t !== contactType)!;
        let threwDifferentType = false;
        try {
          service.createContact({ name: variantName, contactType: otherType });
        } catch (_e) {
          threwDifferentType = true;
        }
        assertEquals(
          threwDifferentType,
          false,
          "Same name with different type should be allowed",
        );

        db.close();
      },
    ),
    { numRuns: 30 },
  );
});

// ─── Property 3: Status Pipeline Advancement ──────────────────────────────────
/**
 * Property 3: Status Pipeline Advancement
 *
 * Advancing a contact with status S produces next(S). Advancing "collaborating"
 * fails. After advancement, last_updated is updated.
 *
 * **Validates: Requirements 6.1, 6.2, 6.3**
 */
Deno.test("Property 3: Status Pipeline - advancement produces correct next status", async () => {
  const statusTransitions: Record<string, string> = {
    identified: "contacted",
    contacted: "responded",
    responded: "collaborating",
  };

  await fc.assert(
    fc.asyncProperty(
      arbContactName,
      arbContactType,
      arbStatus,
      async (name: string, contactType: ContactType, initialStatus: OutreachStatus) => {
        const db = createTestDb();
        const service = createOutreachService(db);

        // Create contact with explicit status
        const contact = service.createContact({ name, contactType, status: initialStatus });
        const beforeTimestamp = contact.lastUpdated;

        if (initialStatus === "collaborating") {
          // Advancing from "collaborating" should fail
          let threw = false;
          try {
            service.advanceStatus(contact.id);
          } catch (_e) {
            threw = true;
          }
          assertEquals(threw, true, "Advancing from 'collaborating' should throw");

          // Status should remain unchanged
          const afterFail = service.getContact(contact.id);
          assertEquals(afterFail!.status, "collaborating");
        } else {
          // Advancing should produce the next status
          const advanced = service.advanceStatus(contact.id);
          assertEquals(
            advanced.status,
            statusTransitions[initialStatus],
            `Advancing from '${initialStatus}' should produce '${statusTransitions[initialStatus]}'`,
          );

          // last_updated should be updated (>= before)
          assertEquals(
            advanced.lastUpdated >= beforeTimestamp,
            true,
            "last_updated should be updated after advancement",
          );
        }

        db.close();
      },
    ),
    { numRuns: 30 },
  );
});

// ─── Property 4: Default Status Invariant ─────────────────────────────────────
/**
 * Property 4: Default Status Invariant
 *
 * Contacts created without explicit status have status "identified".
 *
 * **Validates: Requirements 1.3**
 */
Deno.test("Property 4: Default Status - contacts without explicit status default to 'identified'", async () => {
  await fc.assert(
    fc.asyncProperty(
      arbContactName,
      arbContactType,
      async (name: string, contactType: ContactType) => {
        const db = createTestDb();
        const service = createOutreachService(db);

        // Create without specifying status
        const contact = service.createContact({ name, contactType });
        assertEquals(
          contact.status,
          "identified",
          "Default status should be 'identified'",
        );

        // Verify via getContact as well
        const read = service.getContact(contact.id);
        assertEquals(read!.status, "identified");

        db.close();
      },
    ),
    { numRuns: 30 },
  );
});

// ─── Property 5: Contact List Ordering ────────────────────────────────────────
/**
 * Property 5: Contact List Ordering
 *
 * listContacts() is sorted by last_updated DESC. Adjacent items satisfy ordering.
 *
 * **Validates: Requirements 3.4**
 */
Deno.test("Property 5: Contact List Ordering - listContacts returns results sorted by last_updated DESC", async () => {
  await fc.assert(
    fc.asyncProperty(
      fc.integer({ min: 3, max: 8 }),
      async (count: number) => {
        const db = createTestDb();
        const service = createOutreachService(db);

        // Create multiple contacts with unique names
        for (let i = 0; i < count; i++) {
          service.createContact({
            name: `Contact-${i}-${crypto.randomUUID().slice(0, 6)}`,
            contactType: CONTACT_TYPES[i % CONTACT_TYPES.length]!,
          });
        }

        // Update some contacts to vary their last_updated timestamps
        const allContacts = service.listContacts();
        if (allContacts.length >= 2) {
          const lastContact = allContacts[allContacts.length - 1]!;
          service.updateContact(lastContact.id, {
            name: lastContact.name,
            contactType: lastContact.contactType,
            status: "contacted",
          });
        }

        // Retrieve the list and verify ordering
        const list = service.listContacts();
        for (let i = 0; i < list.length - 1; i++) {
          assertEquals(
            list[i]!.lastUpdated >= list[i + 1]!.lastUpdated,
            true,
            `Contact at index ${i} (${list[i]!.lastUpdated}) should have last_updated >= contact at index ${i + 1} (${list[i + 1]!.lastUpdated})`,
          );
        }

        db.close();
      },
    ),
    { numRuns: 25 },
  );
});

// ─── Property 6: Filter Correctness ──────────────────────────────────────────
/**
 * Property 6: Filter Correctness
 *
 * Filtering by status S returns only records with status S. Filtering by type T
 * returns only records with type T. The filtered result is a subset of the
 * unfiltered result.
 *
 * **Validates: Requirements 3.2, 3.3**
 */
Deno.test("Property 6: Filter Correctness - status and type filters return correct subsets", async () => {
  await fc.assert(
    fc.asyncProperty(
      fc.array(
        fc.record({
          contactType: arbContactType,
          status: arbStatus,
        }),
        { minLength: 2, maxLength: 10 },
      ),
      arbStatus,
      arbContactType,
      async (
        contactSpecs: Array<{ contactType: ContactType; status: OutreachStatus }>,
        filterStatus: OutreachStatus,
        filterType: ContactType,
      ) => {
        const db = createTestDb();
        const service = createOutreachService(db);

        // Create contacts with unique names
        const created: number[] = [];
        for (let i = 0; i < contactSpecs.length; i++) {
          const spec = contactSpecs[i]!;
          try {
            const record = service.createContact({
              name: `Filter-${i}-${crypto.randomUUID().slice(0, 6)}`,
              contactType: spec.contactType,
              status: spec.status,
            });
            created.push(record.id);
          } catch (_e) {
            // Skip duplicates
          }
        }

        if (created.length === 0) {
          db.close();
          return;
        }

        const allContacts = service.listContacts();

        // Filter by status
        const filteredByStatus = service.listContacts({ status: filterStatus });
        for (const contact of filteredByStatus) {
          assertEquals(
            contact.status,
            filterStatus,
            `All contacts filtered by status '${filterStatus}' should have that status`,
          );
        }
        const manualStatusFilter = allContacts.filter((c) => c.status === filterStatus);
        assertEquals(filteredByStatus.length, manualStatusFilter.length);

        // Filter by type
        const filteredByType = service.listContacts({ contactType: filterType });
        for (const contact of filteredByType) {
          assertEquals(
            contact.contactType,
            filterType,
            `All contacts filtered by type '${filterType}' should have that type`,
          );
        }
        const manualTypeFilter = allContacts.filter((c) => c.contactType === filterType);
        assertEquals(filteredByType.length, manualTypeFilter.length);

        db.close();
      },
    ),
    { numRuns: 25 },
  );
});

// ─── Property 7: Public View Shows Only Collaborators ─────────────────────────
/**
 * Property 7: Public View Shows Only Collaborators
 *
 * getCollaborators() returns only records with status "collaborating". Every
 * record in the result has status "collaborating". No record with a different
 * status appears in the result.
 *
 * **Validates: Requirements 8.1, 8.3**
 */
Deno.test("Property 7: Public View - getCollaborators returns only 'collaborating' records", async () => {
  await fc.assert(
    fc.asyncProperty(
      fc.array(
        fc.record({
          contactType: arbContactType,
          status: arbStatus,
        }),
        { minLength: 3, maxLength: 12 },
      ),
      async (contactSpecs: Array<{ contactType: ContactType; status: OutreachStatus }>) => {
        const db = createTestDb();
        const service = createOutreachService(db);

        // Create contacts with unique names and various statuses
        let collaboratingCount = 0;
        for (let i = 0; i < contactSpecs.length; i++) {
          const spec = contactSpecs[i]!;
          try {
            service.createContact({
              name: `Collab-Test-${i}-${crypto.randomUUID().slice(0, 6)}`,
              contactType: spec.contactType,
              status: spec.status,
            });
            if (spec.status === "collaborating") {
              collaboratingCount++;
            }
          } catch (_e) {
            // Skip duplicates
          }
        }

        // Get collaborators
        const collaborators = service.getCollaborators();

        // Every returned record must have status "collaborating"
        for (const contact of collaborators) {
          assertEquals(
            contact.status,
            "collaborating",
            "getCollaborators should only return records with status 'collaborating'",
          );
        }

        // Count should match the number of collaborating contacts we created
        assertEquals(
          collaborators.length,
          collaboratingCount,
          "getCollaborators count should match number of 'collaborating' records",
        );

        // Cross-check: no non-collaborating record should appear
        const allContacts = service.listContacts();
        const nonCollaborating = allContacts.filter((c) => c.status !== "collaborating");
        const collaboratorIds = new Set(collaborators.map((c) => c.id));
        for (const nc of nonCollaborating) {
          assertEquals(
            collaboratorIds.has(nc.id),
            false,
            "Non-collaborating records should not appear in getCollaborators",
          );
        }

        db.close();
      },
    ),
    { numRuns: 25 },
  );
});

// ─── Property 8: Statistics Consistency ───────────────────────────────────────
/**
 * Property 8: Statistics Consistency
 *
 * Sum of byStatus counts equals total. Sum of byType counts equals total.
 * Total equals listContacts().length.
 *
 * **Validates: Requirements 10.1, 10.2, 10.3**
 */
Deno.test("Property 8: Statistics Consistency - sums match total and list length", async () => {
  await fc.assert(
    fc.asyncProperty(
      fc.array(
        fc.record({
          contactType: arbContactType,
          status: arbStatus,
        }),
        { minLength: 0, maxLength: 15 },
      ),
      async (contactSpecs: Array<{ contactType: ContactType; status: OutreachStatus }>) => {
        const db = createTestDb();
        const service = createOutreachService(db);

        // Create contacts with unique names
        for (let i = 0; i < contactSpecs.length; i++) {
          const spec = contactSpecs[i]!;
          try {
            service.createContact({
              name: `Stats-${i}-${crypto.randomUUID().slice(0, 6)}`,
              contactType: spec.contactType,
              status: spec.status,
            });
          } catch (_e) {
            // Skip duplicates
          }
        }

        const stats = service.getStats();
        const allContacts = service.listContacts();

        // Total equals listContacts().length
        assertEquals(
          stats.total,
          allContacts.length,
          "stats.total should equal listContacts().length",
        );

        // Sum of byStatus counts equals total
        const statusSum = Object.values(stats.byStatus).reduce(
          (a: number, b: number) => a + b,
          0,
        );
        assertEquals(
          statusSum,
          stats.total,
          "Sum of byStatus counts should equal total",
        );

        // Sum of byType counts equals total
        const typeSum = Object.values(stats.byType).reduce(
          (a: number, b: number) => a + b,
          0,
        );
        assertEquals(
          typeSum,
          stats.total,
          "Sum of byType counts should equal total",
        );

        // Verify individual status counts match actual filtered counts
        for (const status of STATUSES) {
          const filtered = service.listContacts({ status });
          assertEquals(
            stats.byStatus[status],
            filtered.length,
            `byStatus['${status}'] should match filtered count`,
          );
        }

        // Verify individual type counts match actual filtered counts
        for (const cType of CONTACT_TYPES) {
          const filtered = service.listContacts({ contactType: cType });
          assertEquals(
            stats.byType[cType],
            filtered.length,
            `byType['${cType}'] should match filtered count`,
          );
        }

        db.close();
      },
    ),
    { numRuns: 25 },
  );
});


// ═══════════════════════════════════════════════════════════════════════════════
// Integration Tests — Deterministic scenario-based tests
// ═══════════════════════════════════════════════════════════════════════════════

import {
  assertThrows,
  assertNotEquals,
} from "https://deno.land/std@0.224.0/assert/mod.ts";

// ─── Integration Test: CRUD Operations ────────────────────────────────────────
/**
 * Test CRUD operations with valid data (create, read, update, delete a specific contact).
 *
 * **Validates: Requirements 1.1, 2.1**
 */
Deno.test("Integration: CRUD operations with valid data", () => {
  const db = createTestDb();
  const service = createOutreachService(db);

  // CREATE
  const contact = service.createContact({
    name: "AMSAT-UK",
    contactType: "organisation",
    contactUrl: "https://amsat-uk.org",
    callsign: "G4SWX",
    email: "info@amsat-uk.org",
    notes: "National AMSAT organisation in the UK",
    contactedBy: "VK5DGR",
    contactedDate: "2024-03-15",
  });

  assertEquals(contact.name, "AMSAT-UK");
  assertEquals(contact.contactType, "organisation");
  assertEquals(contact.status, "identified");
  assertEquals(contact.contactUrl, "https://amsat-uk.org");
  assertEquals(contact.callsign, "G4SWX");
  assertEquals(contact.email, "info@amsat-uk.org");
  assertEquals(contact.notes, "National AMSAT organisation in the UK");
  assertEquals(contact.contactedBy, "VK5DGR");
  assertEquals(contact.contactedDate, "2024-03-15");
  assertNotEquals(contact.id, undefined);
  assertNotEquals(contact.createdAt, undefined);
  assertNotEquals(contact.lastUpdated, undefined);

  // READ
  const read = service.getContact(contact.id);
  assertEquals(read !== null, true);
  assertEquals(read!.id, contact.id);
  assertEquals(read!.name, "AMSAT-UK");
  assertEquals(read!.contactType, "organisation");
  assertEquals(read!.email, "info@amsat-uk.org");

  // UPDATE
  const updated = service.updateContact(contact.id, {
    name: "AMSAT-UK (Updated)",
    contactType: "organisation",
    status: "contacted",
    contactUrl: "https://amsat-uk.org/contact",
    email: "outreach@amsat-uk.org",
    notes: "Contacted via email, awaiting response",
    contactedBy: "VK5DGR",
    contactedDate: "2024-03-20",
  });

  assertEquals(updated.name, "AMSAT-UK (Updated)");
  assertEquals(updated.status, "contacted");
  assertEquals(updated.email, "outreach@amsat-uk.org");
  assertEquals(updated.contactUrl, "https://amsat-uk.org/contact");

  // Verify update persisted
  const readAfterUpdate = service.getContact(contact.id);
  assertEquals(readAfterUpdate!.name, "AMSAT-UK (Updated)");
  assertEquals(readAfterUpdate!.status, "contacted");

  // DELETE
  service.deleteContact(contact.id);
  const readAfterDelete = service.getContact(contact.id);
  assertEquals(readAfterDelete, null);

  db.close();
});

// ─── Integration Test: Duplicate Detection (Case-Insensitive) ─────────────────
/**
 * Test duplicate detection with case-insensitive matching using specific examples.
 *
 * **Validates: Requirements 1.4, 2.2**
 */
Deno.test("Integration: Duplicate detection is case-insensitive", () => {
  const db = createTestDb();
  const service = createOutreachService(db);

  // Create initial contact
  service.createContact({
    name: "AMSAT-DL",
    contactType: "organisation",
  });

  // Exact same name and type should throw
  assertThrows(
    () => service.createContact({ name: "AMSAT-DL", contactType: "organisation" }),
    Error,
    "already exists",
  );

  // Lowercase variant should throw
  assertThrows(
    () => service.createContact({ name: "amsat-dl", contactType: "organisation" }),
    Error,
    "already exists",
  );

  // Mixed case variant should throw
  assertThrows(
    () => service.createContact({ name: "Amsat-Dl", contactType: "organisation" }),
    Error,
    "already exists",
  );

  // Same name but different type should succeed
  const differentType = service.createContact({
    name: "AMSAT-DL",
    contactType: "mailing_list",
  });
  assertEquals(differentType.name, "AMSAT-DL");
  assertEquals(differentType.contactType, "mailing_list");

  // isDuplicate should detect case-insensitive matches
  assertEquals(service.isDuplicate("amsat-dl", "organisation"), true);
  assertEquals(service.isDuplicate("AMSAT-DL", "organisation"), true);
  assertEquals(service.isDuplicate("amsat-dl", "club"), false);

  db.close();
});

// ─── Integration Test: Status Pipeline Full Advancement ───────────────────────
/**
 * Test status advancement through the full pipeline:
 * identified → contacted → responded → collaborating
 *
 * **Validates: Requirements 6.1, 6.2**
 */
Deno.test("Integration: Status advancement through full pipeline", () => {
  const db = createTestDb();
  const service = createOutreachService(db);

  const contact = service.createContact({
    name: "University of Adelaide Space Club",
    contactType: "university",
  });

  // Default status is "identified"
  assertEquals(contact.status, "identified");

  // Advance: identified → contacted
  const step1 = service.advanceStatus(contact.id);
  assertEquals(step1.status, "contacted");

  // Advance: contacted → responded
  const step2 = service.advanceStatus(contact.id);
  assertEquals(step2.status, "responded");

  // Advance: responded → collaborating
  const step3 = service.advanceStatus(contact.id);
  assertEquals(step3.status, "collaborating");

  // Verify final state
  const final = service.getContact(contact.id);
  assertEquals(final!.status, "collaborating");

  db.close();
});

// ─── Integration Test: Advancement Failure at Collaborating ───────────────────
/**
 * Test that advancing a contact already at "collaborating" throws an error.
 *
 * **Validates: Requirements 6.2**
 */
Deno.test("Integration: Advancement fails at collaborating status", () => {
  const db = createTestDb();
  const service = createOutreachService(db);

  const contact = service.createContact({
    name: "CubeSat Team Alpha",
    contactType: "cubesat_team",
    status: "collaborating",
  });

  assertEquals(contact.status, "collaborating");

  // Attempting to advance should throw
  assertThrows(
    () => service.advanceStatus(contact.id),
    Error,
    "Cannot advance status",
  );

  // Status should remain unchanged
  const afterAttempt = service.getContact(contact.id);
  assertEquals(afterAttempt!.status, "collaborating");

  db.close();
});

// ─── Integration Test: Filtering by Status and Type ───────────────────────────
/**
 * Test filtering contacts by status and type with known data.
 *
 * **Validates: Requirements 3.2, 3.3**
 */
Deno.test("Integration: Filtering by status and type", () => {
  const db = createTestDb();
  const service = createOutreachService(db);

  // Create contacts with known statuses and types
  service.createContact({ name: "Club Alpha", contactType: "club", status: "identified" });
  service.createContact({ name: "Club Beta", contactType: "club", status: "contacted" });
  service.createContact({ name: "Uni Gamma", contactType: "university", status: "contacted" });
  service.createContact({ name: "Uni Delta", contactType: "university", status: "collaborating" });
  service.createContact({ name: "Individual Epsilon", contactType: "individual", status: "responded" });

  // Filter by status
  const contacted = service.listContacts({ status: "contacted" });
  assertEquals(contacted.length, 2);
  for (const c of contacted) {
    assertEquals(c.status, "contacted");
  }

  const collaborating = service.listContacts({ status: "collaborating" });
  assertEquals(collaborating.length, 1);
  assertEquals(collaborating[0]!.name, "Uni Delta");

  // Filter by type
  const clubs = service.listContacts({ contactType: "club" });
  assertEquals(clubs.length, 2);
  for (const c of clubs) {
    assertEquals(c.contactType, "club");
  }

  const universities = service.listContacts({ contactType: "university" });
  assertEquals(universities.length, 2);
  for (const c of universities) {
    assertEquals(c.contactType, "university");
  }

  // Filter by both status and type
  const contactedClubs = service.listContacts({ status: "contacted", contactType: "club" });
  assertEquals(contactedClubs.length, 1);
  assertEquals(contactedClubs[0]!.name, "Club Beta");

  // No results for non-matching filter
  const noResults = service.listContacts({ status: "collaborating", contactType: "club" });
  assertEquals(noResults.length, 0);

  db.close();
});

// ─── Integration Test: Statistics Calculation ─────────────────────────────────
/**
 * Test statistics calculation with known data.
 *
 * **Validates: Requirements 10.1**
 */
Deno.test("Integration: Statistics calculation with known data", () => {
  const db = createTestDb();
  const service = createOutreachService(db);

  // Create contacts with known distribution
  service.createContact({ name: "Contact A", contactType: "club", status: "identified" });
  service.createContact({ name: "Contact B", contactType: "club", status: "identified" });
  service.createContact({ name: "Contact C", contactType: "university", status: "contacted" });
  service.createContact({ name: "Contact D", contactType: "individual", status: "responded" });
  service.createContact({ name: "Contact E", contactType: "organisation", status: "collaborating" });
  service.createContact({ name: "Contact F", contactType: "organisation", status: "collaborating" });

  const stats = service.getStats();

  // Total
  assertEquals(stats.total, 6);

  // By status
  assertEquals(stats.byStatus.identified, 2);
  assertEquals(stats.byStatus.contacted, 1);
  assertEquals(stats.byStatus.responded, 1);
  assertEquals(stats.byStatus.collaborating, 2);

  // By type
  assertEquals(stats.byType.club, 2);
  assertEquals(stats.byType.university, 1);
  assertEquals(stats.byType.individual, 1);
  assertEquals(stats.byType.organisation, 2);
  assertEquals(stats.byType.facebook_group, 0);
  assertEquals(stats.byType.mailing_list, 0);
  assertEquals(stats.byType.cubesat_team, 0);

  // Sum invariants
  const statusSum = Object.values(stats.byStatus).reduce((a, b) => a + b, 0);
  assertEquals(statusSum, stats.total);

  const typeSum = Object.values(stats.byType).reduce((a, b) => a + b, 0);
  assertEquals(typeSum, stats.total);

  db.close();
});

// ─── Integration Test: getCollaborators Returns Only Collaborating Records ────
/**
 * Test that getCollaborators returns only records with "collaborating" status.
 *
 * **Validates: Requirements 8.1**
 */
Deno.test("Integration: getCollaborators returns only collaborating records", () => {
  const db = createTestDb();
  const service = createOutreachService(db);

  // Create contacts at various statuses
  service.createContact({ name: "Identified Group", contactType: "facebook_group", status: "identified" });
  service.createContact({ name: "Contacted List", contactType: "mailing_list", status: "contacted" });
  service.createContact({ name: "Responded Club", contactType: "club", status: "responded" });
  service.createContact({ name: "Collaborating Uni", contactType: "university", status: "collaborating" });
  service.createContact({ name: "Collaborating Team", contactType: "cubesat_team", status: "collaborating" });
  service.createContact({ name: "Another Identified", contactType: "individual", status: "identified" });

  const collaborators = service.getCollaborators();

  // Should return exactly 2 collaborating records
  assertEquals(collaborators.length, 2);

  // All returned records should have "collaborating" status
  for (const c of collaborators) {
    assertEquals(c.status, "collaborating");
  }

  // Verify the correct records are returned
  const names = collaborators.map((c) => c.name).sort();
  assertEquals(names, ["Collaborating Team", "Collaborating Uni"]);

  db.close();
});

// ─── Integration Test: Default Status ─────────────────────────────────────────
/**
 * Test that contacts created without explicit status default to "identified".
 *
 * **Validates: Requirements 1.3**
 */
Deno.test("Integration: Default status is 'identified' when not specified", () => {
  const db = createTestDb();
  const service = createOutreachService(db);

  const contact = service.createContact({
    name: "New Contact Without Status",
    contactType: "mailing_list",
  });

  assertEquals(contact.status, "identified");

  // Verify via getContact
  const read = service.getContact(contact.id);
  assertEquals(read!.status, "identified");

  db.close();
});

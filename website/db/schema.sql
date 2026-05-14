CREATE TABLE IF NOT EXISTS contact_submissions (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  callsign_or_org TEXT,
  area_of_interest TEXT,
  message TEXT NOT NULL,
  submitted_at TEXT DEFAULT (datetime('now'))
);

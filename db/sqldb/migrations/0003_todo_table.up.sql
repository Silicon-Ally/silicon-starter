BEGIN;

CREATE TABLE task (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  body TEXT NOT NULL,
  tags TEXT NOT NULL,
  created_by TEXT NOT NULL REFERENCES user_account(id)
);

COMMIT;
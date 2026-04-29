-- Allow accounts to be created automatically on statement upload.
-- user_id becomes nullable (no auth context during ingest).
-- A unique constraint on (bank_name, account_number) enables upsert.

BEGIN;

ALTER TABLE accounts ALTER COLUMN user_id DROP NOT NULL;

ALTER TABLE accounts
    ADD CONSTRAINT accounts_bank_account_unique UNIQUE (bank_name, account_number);

COMMIT;

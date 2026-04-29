-- Change accounts.id and transactions.account_id from bigint to text.
-- Existing integer values are cast to their text representation.

BEGIN;

-- Drop the FK constraint and index that depend on account_id type
ALTER TABLE transactions DROP CONSTRAINT IF EXISTS transactions_account_id_fkey;
DROP INDEX IF EXISTS idx_transactions_account_date;

-- Change accounts.id: drop identity and alter type to text
ALTER TABLE accounts ALTER COLUMN id DROP IDENTITY IF EXISTS;
ALTER TABLE accounts ALTER COLUMN id TYPE text USING id::text;

-- Change transactions.account_id to text
ALTER TABLE transactions ALTER COLUMN account_id TYPE text USING account_id::text;

-- Restore FK constraint and index
ALTER TABLE transactions
    ADD CONSTRAINT transactions_account_id_fkey
    FOREIGN KEY (account_id) REFERENCES accounts (id) ON DELETE CASCADE;

CREATE INDEX idx_transactions_account_date ON transactions (account_id, date);

COMMIT;

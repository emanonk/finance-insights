-- Prevent duplicate transactions on ingest.
-- Two rows are considered duplicates when every business-key column matches.
-- NULL values are treated as equal for this purpose via COALESCE.

BEGIN;

CREATE UNIQUE INDEX transactions_dedup_idx ON transactions (
    account_id,
    date,
    COALESCE(bank_reference, ''),
    COALESCE(transaction_reference, ''),
    COALESCE(merchant_identifier, ''),
    amount
);

COMMIT;

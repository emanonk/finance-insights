-- Finance Insights — reporting demo seed
-- Anchor: transactions from 2024-01-01 through 2025-04-30 (reproducible screenshots).
-- Expect 001_schema.sql applied first on an empty database.

BEGIN;

INSERT INTO users (name, username, password)
VALUES ('Demo User', 'demo', '$2a$10$placeholder_hash_demo_only');

INSERT INTO accounts (user_id, bank_name, account_number, currency)
VALUES (
    (SELECT id FROM users WHERE username = 'demo'),
    'Piraeus',
    'GR** **** **** 1234',
    'EUR'
);

-- Primary tags (spec list)
INSERT INTO tags (name, type) VALUES
    ('Food', 'primary'),
    ('Transport', 'primary'),
    ('Home', 'primary'),
    ('Family', 'primary'),
    ('Personal', 'primary'),
    ('Health', 'primary'),
    ('Finance', 'primary'),
    ('Leisure', 'primary'),
    ('Travel', 'primary'),
    ('Income', 'primary'),
    ('Transfers', 'primary'),
    ('Miscellaneous', 'primary'),
    ('SpecialCase', 'primary');

-- Secondary tags (examples)
INSERT INTO tags (name, type) VALUES
    ('groceries', 'secondary'),
    ('eating out', 'secondary'),
    ('fuel', 'secondary'),
    ('parking', 'secondary'),
    ('electricity', 'secondary'),
    ('rent', 'secondary'),
    ('pharmacy', 'secondary'),
    ('haircut', 'secondary'),
    ('netflix', 'secondary'),
    ('toys', 'secondary'),
    ('school', 'secondary'),
    ('coffee', 'secondary'),
    ('subscription', 'secondary');

-- Merchants + secondary tag links
INSERT INTO merchants (identifier_name, primary_tag_id, default_title) VALUES
    ('ab_supermarket', (SELECT id FROM tags WHERE name = 'Food' AND type = 'primary'), 'AB Vasilopoulos'),
    ('starbucks', (SELECT id FROM tags WHERE name = 'Food' AND type = 'primary'), 'Starbucks'),
    ('efood', (SELECT id FROM tags WHERE name = 'Food' AND type = 'primary'), 'eFood'),
    ('shell', (SELECT id FROM tags WHERE name = 'Transport' AND type = 'primary'), 'Shell'),
    ('parking_athens', (SELECT id FROM tags WHERE name = 'Transport' AND type = 'primary'), 'Municipal Parking'),
    ('dei', (SELECT id FROM tags WHERE name = 'Home' AND type = 'primary'), 'DEI Electricity'),
    ('rent_landlord', (SELECT id FROM tags WHERE name = 'Home' AND type = 'primary'), 'Rent'),
    ('insurance_acme', (SELECT id FROM tags WHERE name = 'Finance' AND type = 'primary'), 'Acme Insurance'),
    ('gym_fitness', (SELECT id FROM tags WHERE name = 'Health' AND type = 'primary'), 'Fit Gym'),
    ('pharmacy_chain', (SELECT id FROM tags WHERE name = 'Health' AND type = 'primary'), 'Pharmacy'),
    ('netflix', (SELECT id FROM tags WHERE name = 'Leisure' AND type = 'primary'), 'Netflix'),
    ('salary_employer', (SELECT id FROM tags WHERE name = 'Income' AND type = 'primary'), 'Salary');

INSERT INTO merchant_secondary_tags (merchant_id, tag_id)
SELECT m.id, t.id
FROM merchants m
JOIN tags t ON t.type = 'secondary' AND (
    (m.identifier_name = 'ab_supermarket' AND t.name = 'groceries')
    OR (m.identifier_name = 'starbucks' AND t.name = 'coffee')
    OR (m.identifier_name = 'efood' AND t.name = 'eating out')
    OR (m.identifier_name = 'shell' AND t.name = 'fuel')
    OR (m.identifier_name = 'parking_athens' AND t.name = 'parking')
    OR (m.identifier_name = 'dei' AND t.name = 'electricity')
    OR (m.identifier_name = 'rent_landlord' AND t.name = 'rent')
    OR (m.identifier_name = 'insurance_acme' AND t.name = 'subscription')
    OR (m.identifier_name = 'gym_fitness' AND t.name = 'subscription')
    OR (m.identifier_name = 'pharmacy_chain' AND t.name = 'pharmacy')
    OR (m.identifier_name = 'netflix' AND t.name = 'netflix')
);

-- Helper: single demo account
-- Recurring-style fixed debits (16 monthly hits each; >=10 per calendar year for recurring report)
INSERT INTO transactions (
    account_id, date, merchant_identifier, direction, amount, amount1,
    description, payment_method, statement_file_name, balance_after_transaction,
    bank_reference_number, reference
)
SELECT
    a.id,
    ('2024-01-01'::date + (mm * interval '1 month'))::date,
    'rent_landlord',
    'debit',
    850.00,
    850.00,
    'RENT PAYMENT',
    'transfer',
    format('%s-piraeus-%s.pdf', extract(year FROM ('2024-01-01'::date + (mm * interval '1 month')))::int, to_char('2024-01-01'::date + (mm * interval '1 month'), 'YYYY-MM')),
    NULL::numeric,
    'REF-' || to_char('2024-01-01'::date + (mm * interval '1 month'), 'YYYYMMDD') || '-rent',
    'REF-' || to_char('2024-01-01'::date + (mm * interval '1 month'), 'YYYYMMDD')
FROM accounts a
CROSS JOIN generate_series(0, 15) AS mm;

INSERT INTO transactions (
    account_id, date, merchant_identifier, direction, amount, amount1,
    description, payment_method, statement_file_name, balance_after_transaction,
    bank_reference_number, reference
)
SELECT
    a.id,
    ('2024-01-03'::date + (mm * interval '1 month'))::date,
    'insurance_acme',
    'debit',
    45.50,
    45.50,
    'ACME INSURANCE PREMIUM',
    'card',
    format('%s-piraeus-%s.pdf', extract(year FROM ('2024-01-03'::date + (mm * interval '1 month')))::int, to_char('2024-01-03'::date + (mm * interval '1 month'), 'YYYY-MM')),
    NULL::numeric,
    'REF-' || to_char('2024-01-03'::date + (mm * interval '1 month'), 'YYYYMMDD') || '-ins',
    'REF-' || to_char('2024-01-03'::date + (mm * interval '1 month'), 'YYYYMMDD')
FROM accounts a
CROSS JOIN generate_series(0, 15) AS mm;

INSERT INTO transactions (
    account_id, date, merchant_identifier, direction, amount, amount1,
    description, payment_method, statement_file_name, balance_after_transaction,
    bank_reference_number, reference
)
SELECT
    a.id,
    ('2024-01-05'::date + (mm * interval '1 month'))::date,
    'netflix',
    'debit',
    13.99,
    13.99,
    'NETFLIX.COM',
    'card',
    format('%s-piraeus-%s.pdf', extract(year FROM ('2024-01-05'::date + (mm * interval '1 month')))::int, to_char('2024-01-05'::date + (mm * interval '1 month'), 'YYYY-MM')),
    NULL::numeric,
    'REF-' || to_char('2024-01-05'::date + (mm * interval '1 month'), 'YYYYMMDD') || '-nf',
    'REF-' || to_char('2024-01-05'::date + (mm * interval '1 month'), 'YYYYMMDD')
FROM accounts a
CROSS JOIN generate_series(0, 15) AS mm;

INSERT INTO transactions (
    account_id, date, merchant_identifier, direction, amount, amount1,
    description, payment_method, statement_file_name, balance_after_transaction,
    bank_reference_number, reference
)
SELECT
    a.id,
    ('2024-01-07'::date + (mm * interval '1 month'))::date,
    'gym_fitness',
    'debit',
    39.00,
    39.00,
    'FIT GYM MEMBERSHIP',
    'card',
    format('%s-piraeus-%s.pdf', extract(year FROM ('2024-01-07'::date + (mm * interval '1 month')))::int, to_char('2024-01-07'::date + (mm * interval '1 month'), 'YYYY-MM')),
    NULL::numeric,
    'REF-' || to_char('2024-01-07'::date + (mm * interval '1 month'), 'YYYYMMDD') || '-gym',
    'REF-' || to_char('2024-01-07'::date + (mm * interval '1 month'), 'YYYYMMDD')
FROM accounts a
CROSS JOIN generate_series(0, 15) AS mm;

INSERT INTO transactions (
    account_id, date, merchant_identifier, direction, amount, amount1,
    description, payment_method, statement_file_name, balance_after_transaction,
    bank_reference_number, reference
)
SELECT
    a.id,
    ('2024-01-10'::date + (mm * interval '1 month'))::date,
    'dei',
    'debit',
    62.00,
    62.00,
    'DEI S.A. BILL',
    'card',
    format('%s-piraeus-%s.pdf', extract(year FROM ('2024-01-10'::date + (mm * interval '1 month')))::int, to_char('2024-01-10'::date + (mm * interval '1 month'), 'YYYY-MM')),
    NULL::numeric,
    'REF-' || to_char('2024-01-10'::date + (mm * interval '1 month'), 'YYYYMMDD') || '-dei',
    'REF-' || to_char('2024-01-10'::date + (mm * interval '1 month'), 'YYYYMMDD')
FROM accounts a
CROSS JOIN generate_series(0, 15) AS mm;

-- Salary credits (monthly)
INSERT INTO transactions (
    account_id, date, merchant_identifier, direction, amount, amount1,
    description, payment_method, statement_file_name,
    bank_reference_number, reference
)
SELECT
    a.id,
    ('2024-01-25'::date + (m * interval '1 month'))::date,
    'salary_employer',
    'credit',
    2500.00,
    2500.00,
    'SALARY ACME CORP',
    'transfer',
    format(
        '%s-piraeus-%s.pdf',
        extract(year FROM ('2024-01-25'::date + (m * interval '1 month')))::int,
        to_char('2024-01-25'::date + (m * interval '1 month'), 'YYYY-MM')
    ),
    'SAL-' || to_char('2024-01-25'::date + (m * interval '1 month'), 'YYYYMM'),
    'SAL-' || to_char('2024-01-25'::date + (m * interval '1 month'), 'YYYYMM')
FROM accounts a
CROSS JOIN generate_series(0, 15) AS m;

-- Groceries: ~100 debits, pseudo-random amounts, Tue/Sat pattern
INSERT INTO transactions (
    account_id, date, merchant_identifier, direction, amount, amount1,
    description, payment_method, statement_file_name,
    bank_reference_number, reference
)
SELECT
    a.id,
    d,
    'ab_supermarket',
    'debit',
    (18.00 + (random() * 85)::numeric(14, 2))::numeric(14, 2),
    NULL,
    'AB VASILOPOULOS S.A.',
    'card',
    format('%s-piraeus-%s.pdf', extract(year FROM d)::int, to_char(d, 'YYYY-MM')),
    'AB-' || to_char(d, 'YYYYMMDD') || '-' || n,
    'AB-' || to_char(d, 'YYYYMMDD') || '-' || n
FROM accounts a
CROSS JOIN LATERAL (
    SELECT
        ('2024-01-02'::date + (n * 5 + (n % 3)))::date AS d,
        n
    FROM generate_series(1, 100) AS n
) AS g;

-- Coffee: ~80 debits
INSERT INTO transactions (
    account_id, date, merchant_identifier, direction, amount, amount1,
    description, payment_method, statement_file_name,
    bank_reference_number, reference
)
SELECT
    a.id,
    ('2024-01-04'::date + ((n - 1) * 6 + (n % 5)) * interval '1 day')::date,
    'starbucks',
    'debit',
    (4.50 + (random() * 6)::numeric(14, 2))::numeric(14, 2),
    NULL,
    'STARBUCKS STORE',
    'card',
    format(
        '%s-piraeus-%s.pdf',
        extract(year FROM ('2024-01-04'::date + ((n - 1) * 6 + (n % 5)) * interval '1 day'))::int,
        to_char('2024-01-04'::date + ((n - 1) * 6 + (n % 5)) * interval '1 day', 'YYYY-MM')
    ),
    'SB-' || n::text,
    'SB-' || n::text
FROM accounts a
CROSS JOIN generate_series(1, 80) AS n;

-- Fuel: ~48 debits
INSERT INTO transactions (
    account_id, date, merchant_identifier, direction, amount, amount1,
    description, payment_method, statement_file_name,
    bank_reference_number, reference
)
SELECT
    a.id,
    ('2024-01-08'::date + ((n - 1) * 10 + (n % 7)) * interval '1 day')::date,
    'shell',
    'debit',
    (35.00 + (random() * 45)::numeric(14, 2))::numeric(14, 2),
    NULL,
    'SHELL FUEL',
    'card',
    format(
        '%s-piraeus-%s.pdf',
        extract(year FROM ('2024-01-08'::date + ((n - 1) * 10 + (n % 7)) * interval '1 day'))::int,
        to_char('2024-01-08'::date + ((n - 1) * 10 + (n % 7)) * interval '1 day', 'YYYY-MM')
    ),
    'SH-' || n::text,
    'SH-' || n::text
FROM accounts a
CROSS JOIN generate_series(1, 48) AS n;

-- eFood: ~32 debits
INSERT INTO transactions (
    account_id, date, merchant_identifier, direction, amount, amount1,
    description, payment_method, statement_file_name,
    bank_reference_number, reference
)
SELECT
    a.id,
    ('2024-01-12'::date + ((n - 1) * 15 + (n % 4)) * interval '1 day')::date,
    'efood',
    'debit',
    (12.00 + (random() * 28)::numeric(14, 2))::numeric(14, 2),
    NULL,
    'EFOOD.GR',
    'card',
    format(
        '%s-piraeus-%s.pdf',
        extract(year FROM ('2024-01-12'::date + ((n - 1) * 15 + (n % 4)) * interval '1 day'))::int,
        to_char('2024-01-12'::date + ((n - 1) * 15 + (n % 4)) * interval '1 day', 'YYYY-MM')
    ),
    'EF-' || n::text,
    'EF-' || n::text
FROM accounts a
CROSS JOIN generate_series(1, 32) AS n;

-- Pharmacy: ~10 debits
INSERT INTO transactions (
    account_id, date, merchant_identifier, direction, amount, amount1,
    description, payment_method, statement_file_name,
    bank_reference_number, reference
)
SELECT
    a.id,
    ('2024-01-20'::date + ((n - 1) * 48) * interval '1 day')::date,
    'pharmacy_chain',
    'debit',
    (6.00 + (random() * 42)::numeric(14, 2))::numeric(14, 2),
    NULL,
    'PHARMACY CHAIN',
    'card',
    format(
        '%s-piraeus-%s.pdf',
        extract(year FROM ('2024-01-20'::date + ((n - 1) * 48) * interval '1 day'))::int,
        to_char('2024-01-20'::date + ((n - 1) * 48) * interval '1 day', 'YYYY-MM')
    ),
    'PH-' || n::text,
    'PH-' || n::text
FROM accounts a
CROSS JOIN generate_series(1, 10) AS n;

-- Parking: ~16 debits
INSERT INTO transactions (
    account_id, date, merchant_identifier, direction, amount, amount1,
    description, payment_method, statement_file_name,
    bank_reference_number, reference
)
SELECT
    a.id,
    ('2024-01-15'::date + ((n - 1) * interval '1 month'))::date,
    'parking_athens',
    'debit',
    (3.00 + (n % 4))::numeric(14, 2),
    NULL,
    'ATHENS MUNICIPAL PARKING',
    'card',
    format(
        '%s-piraeus-%s.pdf',
        extract(year FROM ('2024-01-15'::date + ((n - 1) * interval '1 month')))::int,
        to_char('2024-01-15'::date + ((n - 1) * interval '1 month'), 'YYYY-MM')
    ),
    'PK-' || n::text,
    'PK-' || n::text
FROM accounts a
CROSS JOIN generate_series(1, 16) AS n;

-- One large SpecialCase-style debit (car purchase) — tagged merchant optional; leave merchant NULL for "unrecognized" demo
INSERT INTO transactions (
    account_id, date, merchant_identifier, direction, amount, amount1,
    description, payment_method, statement_file_name,
    bank_reference_number, reference
)
SELECT
    a.id,
    '2024-06-14'::date,
    NULL,
    'debit',
    12500.00,
    12500.00,
    'AUTO DEALER S.A. VEHICLE PURCHASE',
    'transfer',
    '2024-piraeus-2024-06.pdf',
    'CAR-20240614',
    'CAR-20240614'
FROM accounts a;

-- Parser metadata on a sample of rows (every 7th transaction id)
INSERT INTO transaction_extra (transaction_id, note, in_report, parser_name_version)
SELECT t.id, 'Parsed by demo seed', true, 'piraeus_csv/v1'
FROM transactions t
WHERE t.id % 7 = 0;

-- Optional: align balances (simple running total per account by date)
WITH ordered AS (
    SELECT
        id,
        SUM(
            CASE direction
                WHEN 'credit' THEN amount
                ELSE -amount
            END
        ) OVER (
            PARTITION BY account_id
            ORDER BY date, id
            ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
        ) AS bal
    FROM transactions
)
UPDATE transactions t
SET balance_after_transaction = o.bal
FROM ordered o
WHERE t.id = o.id;

SELECT setval(pg_get_serial_sequence('users', 'id'), (SELECT COALESCE(MAX(id), 1) FROM users));
SELECT setval(pg_get_serial_sequence('accounts', 'id'), (SELECT COALESCE(MAX(id), 1) FROM accounts));
SELECT setval(pg_get_serial_sequence('tags', 'id'), (SELECT COALESCE(MAX(id), 1) FROM tags));
SELECT setval(pg_get_serial_sequence('merchants', 'id'), (SELECT COALESCE(MAX(id), 1) FROM merchants));
SELECT setval(pg_get_serial_sequence('transactions', 'id'), (SELECT COALESCE(MAX(id), 1) FROM transactions));
SELECT setval(pg_get_serial_sequence('transaction_extra', 'id'), (SELECT COALESCE(MAX(id), 1) FROM transaction_extra));

COMMIT;

CREATE TABLE IF NOT EXISTS transactions (
    id                        uuid PRIMARY KEY,
    statement_id              uuid NOT NULL REFERENCES statements(id) ON DELETE CASCADE,
    account_id                text,
    date                      date NOT NULL,
    merchant_identifier       text,
    description               text NOT NULL,
    direction                 text NOT NULL CHECK (direction IN ('Debit', 'Credit')),
    amount                    numeric(14, 2) NOT NULL,
    balance_after_transaction numeric(14, 2),
    mcc_code                  text,
    card_masked               text,
    reference                 text,
    bank_reference_number     text,
    payment_method            text,
    created_at                timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS transactions_date_idx ON transactions (date DESC, id DESC);
CREATE INDEX IF NOT EXISTS transactions_statement_id_idx ON transactions (statement_id);

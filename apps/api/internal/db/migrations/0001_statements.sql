CREATE TABLE IF NOT EXISTS statements (
    id          uuid PRIMARY KEY,
    file_name   text NOT NULL,
    stored_path text NOT NULL,
    uploaded_at timestamptz NOT NULL DEFAULT now()
);

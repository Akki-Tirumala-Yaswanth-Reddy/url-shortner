CREATE TABLE IF NOT EXISTS urls (
  id           BIGSERIAL    PRIMARY KEY,
  user_name    TEXT         NOT NULL,
  short_code   TEXT         UNIQUE,
  original_url TEXT         NOT NULL,
  created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_urls_short_code ON urls (short_code);
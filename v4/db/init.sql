CREATE TABLE IF NOT EXISTS urls (
  id               BIGSERIAL    PRIMARY KEY,
  user_name        TEXT         NOT NULL,
  short_code       TEXT         UNIQUE,
  original_url     TEXT         NOT NULL,
  created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  click_count      BIGINT       NOT NULL DEFAULT 0,
  last_accessed_at TIMESTAMPTZ  NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_urls_short_code ON urls (short_code);
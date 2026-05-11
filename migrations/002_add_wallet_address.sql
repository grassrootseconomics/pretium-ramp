ALTER TABLE onramp ADD COLUMN IF NOT EXISTS wallet_address TEXT;
ALTER TABLE offramp ADD COLUMN IF NOT EXISTS wallet_address TEXT;

CREATE INDEX IF NOT EXISTS onramp_wallet_address_idx ON onramp(wallet_address);
CREATE INDEX IF NOT EXISTS offramp_wallet_address_idx ON offramp(wallet_address);

---- create above / drop below ----

DROP INDEX IF EXISTS offramp_wallet_address_idx;
DROP INDEX IF EXISTS onramp_wallet_address_idx;

ALTER TABLE offramp DROP COLUMN IF EXISTS wallet_address;
ALTER TABLE onramp DROP COLUMN IF EXISTS wallet_address;

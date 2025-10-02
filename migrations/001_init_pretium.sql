-- Keystore
CREATE TABLE IF NOT EXISTS non_custodial_link (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    public_key TEXT NOT NULL,
    phone_number TEXT NOT NULL,
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS public_key_idx ON non_custodial_link(public_key);
CREATE INDEX IF NOT EXISTS phone_number_idx ON non_custodial_link(phone_number);

CREATE TABLE IF NOT EXISTS offramp (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    pretium_id TEXT NOT NULL,
    pretium_status TEXT NOT NULL DEFAULT 'PENDING',
    mpesa_confirmation TEXT,
    phone_number TEXT NOT NULL,
    amount_usd NUMERIC NOT NULL,
    amount_kes NUMERIC NOT NULL,
    tx_hash TEXT NOT NULL,
    token_address TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS onramp (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    pretium_id TEXT NOT NULL,
    pretium_status TEXT NOT NULL DEFAULT 'PENDING',
    mpesa_confirmation TEXT,
    phone_number TEXT NOT NULL,
    amount_usd NUMERIC NOT NULL,
    amount_kes NUMERIC NOT NULL,
    tx_hash TEXT NOT NULL,
    token_address TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- updated_at function
create function update_timestamp()
    returns trigger
as $$
begin
    new.updated_at = current_timestamp;
    return new;
end;
$$ language plpgsql;

create trigger update_onramp_timestamp
    before update on onramp
for each row
execute procedure update_timestamp();

create trigger update_offramp_timestamp
    before update on offramp
for each row
execute procedure update_timestamp();

create trigger update_non_custodial_link_timestamp
    before update on non_custodial_link
for each row
execute procedure update_timestamp();
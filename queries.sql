-- Non-custodial link queries

--name: insert-non-custodial-link
-- $1: public_key
-- $2: phone_number
INSERT INTO non_custodial_link(
    public_key,
    phone_number
) VALUES($1, $2) ON CONFLICT DO NOTHING;

--name: get-non-custodial-link-by-public-key
-- $1: public_key
SELECT * FROM non_custodial_link WHERE public_key = $1 AND active = true;

--name: get-non-custodial-link-by-phone
-- $1: phone_number
SELECT * FROM non_custodial_link WHERE phone_number = $1 AND active = true;

--name: deactivate-non-custodial-link
-- $1: phone_number
UPDATE non_custodial_link SET active = false WHERE phone_number = $1;


-- Offramp queries

--name: insert-offramp
-- $1: pretium_id
-- $2: phone_number
-- $3: amount_usd
-- $4: amount_kes
-- $5: tx_hash
-- $6: token_address
INSERT INTO offramp(
    pretium_id,
    phone_number,
    amount_usd,
    amount_kes,
    tx_hash,
    token_address
) VALUES($1, $2, $3, $4, $5, $6);

--name: get-offramp-by-pretium-id
-- $1: pretium_id
SELECT * FROM offramp WHERE pretium_id = $1;

--name: get-offramp-by-tx-hash
-- $1: tx_hash
SELECT * FROM offramp WHERE tx_hash = $1;

--name: get-offramp-by-phone
-- $1: phone_number
SELECT * FROM offramp WHERE phone_number = $1 ORDER BY created_at DESC;

--name: update-offramp-status
-- $1: pretium_status
-- $2: pretium_id
UPDATE offramp SET pretium_status = $1 WHERE pretium_id = $2;

--name: update-offramp-mpesa-confirmation
-- $1: mpesa_confirmation
-- $2: pretium_status
-- $3: pretium_id
UPDATE offramp SET mpesa_confirmation = $1, pretium_status = $2 WHERE pretium_id = $3;

--name: get-stale-offramps
SELECT * FROM offramp 
WHERE mpesa_confirmation IS NULL 
  AND created_at < NOW() - INTERVAL '1 minute' 
ORDER BY created_at ASC 
LIMIT 100;

--name: insert-onramp
-- $1: pretium_id
-- $2: phone_number
-- $3: amount_usd
-- $4: amount_kes
-- $5: tx_hash
-- $6: token_address
INSERT INTO onramp(
    pretium_id,
    phone_number,
    amount_usd,
    amount_kes,
    tx_hash,
    token_address
) VALUES($1, $2, $3, $4, $5, $6);

--name: get-onramp-by-pretium-id
-- $1: pretium_id
SELECT * FROM onramp WHERE pretium_id = $1;

--name: get-onramp-by-tx-hash
-- $1: tx_hash
SELECT * FROM onramp WHERE tx_hash = $1;

--name: get-onramp-by-phone
-- $1: phone_number
SELECT * FROM onramp WHERE phone_number = $1 ORDER BY created_at DESC;

--name: update-onramp-status
-- $1: pretium_status
-- $2: pretium_id
UPDATE onramp SET pretium_status = $1 WHERE pretium_id = $2;

--name: update-onramp-mpesa-confirmation
-- $1: mpesa_confirmation
-- $2: pretium_status
-- $3: pretium_id
UPDATE onramp SET mpesa_confirmation = $1, pretium_status = $2 WHERE pretium_id = $3;

--name: get-stale-onramps
SELECT * FROM onramp 
WHERE mpesa_confirmation IS NULL 
  AND created_at < NOW() - INTERVAL '1 minute' 
ORDER BY created_at ASC 
LIMIT 100;
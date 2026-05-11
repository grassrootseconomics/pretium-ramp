# API Documentation

## Base URL
```
/api/v1
```

## Endpoints

### POST /callback/{callbackSecret}
Pretium webhook for transaction status updates

### POST /trigger-onramp
Initiates fiat-to-crypto onramp
```
{
  "address": "ethereum address",
  "phoneNumber": "phone number (optional)",
  "asset": "USDT | USDC | cUSD",
  "amount": 20-250000
}
```

### POST /link
Links public key to phone number
```
{
  "publicKey": "ethereum address",
  "phoneNumber": "phone number"
}
```

### GET /link/{phoneNumber}
Returns all public keys linked to phone number

### GET /transactions/{phoneNumber}
Returns all transactions for phone number

### GET /transactions-by-address/{address}
Returns all onramps and offramps where the supplied Ethereum address was used (onramp recipient or offramp initiator). Each record includes its `pretium_status` and `pretium_id` (reference). Returns 200 with empty arrays if the address has no recorded transactions. Note: only transactions created after the `wallet_address` column was added are queryable here.

### GET /transactions-recent
Returns recent transactions (last 3 days)

### GET /rates
Returns Pretium rates

### GET /metrics
Prometheus metrics (if enabled)
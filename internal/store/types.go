package store

import "time"

type (
	NonCustodialLink struct {
		ID          int       `db:"id"`
		PublicKey   string    `db:"public_key"`
		PhoneNumber string    `db:"phone_number"`
		Active      bool      `db:"active"`
		CreatedAt   time.Time `db:"created_at"`
		UpdatedAt   time.Time `db:"updated_at"`
	}

	Offramp struct {
		ID                int       `db:"id"`
		PretiumID         string    `db:"pretium_id"`
		PretiumStatus     string    `db:"pretium_status"`
		MpesaConfirmation *string   `db:"mpesa_confirmation"`
		PhoneNumber       string    `db:"phone_number"`
		AmountUSD         string    `db:"amount_usd"`
		AmountKES         string    `db:"amount_kes"`
		TxHash            string    `db:"tx_hash"`
		TokenAddress      string    `db:"token_address"`
		CreatedAt         time.Time `db:"created_at"`
		UpdatedAt         time.Time `db:"updated_at"`
	}

	Onramp struct {
		ID                int       `db:"id"`
		PretiumID         string    `db:"pretium_id"`
		PretiumStatus     string    `db:"pretium_status"`
		MpesaConfirmation *string   `db:"mpesa_confirmation"`
		PhoneNumber       string    `db:"phone_number"`
		AmountUSD         string    `db:"amount_usd"`
		AmountKES         string    `db:"amount_kes"`
		TxHash            string    `db:"tx_hash"`
		TokenAddress      string    `db:"token_address"`
		CreatedAt         time.Time `db:"created_at"`
		UpdatedAt         time.Time `db:"updated_at"`
	}
)

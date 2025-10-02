package store

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store interface {
	Pool() *pgxpool.Pool

	InsertNonCustodialLink(ctx context.Context, publicKey, phoneNumber string) error
	GetNonCustodialLinkByPublicKey(ctx context.Context, publicKey string) (*NonCustodialLink, error)
	GetNonCustodialLinkByPhone(ctx context.Context, phoneNumber string) (*NonCustodialLink, error)
	DeactivateNonCustodialLink(ctx context.Context, phoneNumber string) error

	InsertOfframp(ctx context.Context, pretiumID, phoneNumber, amountUSD, amountKES, txHash, tokenAddress string) (int, error)
	GetOfframpByPretiumID(ctx context.Context, pretiumID string) (*Offramp, error)
	GetOfframpByTxHash(ctx context.Context, txHash string) (*Offramp, error)
	GetOfframpsByPhone(ctx context.Context, phoneNumber string) ([]Offramp, error)
	UpdateOfframpStatus(ctx context.Context, pretiumStatus, pretiumID string) error
	UpdateOfframpMpesaConfirmation(ctx context.Context, mpesaConfirmation, pretiumStatus string, id int) error

	InsertOnramp(ctx context.Context, pretiumID, phoneNumber string, amountUSD, amountKES, txHash, tokenAddress string) (int, error)
	GetOnrampByPretiumID(ctx context.Context, pretiumID string) (*Onramp, error)
	GetOnrampByTxHash(ctx context.Context, txHash string) (*Onramp, error)
	GetOnrampsByPhone(ctx context.Context, phoneNumber string) ([]Onramp, error)
	UpdateOnrampStatus(ctx context.Context, pretiumStatus, pretiumID string) error
	UpdateOnrampMpesaConfirmation(ctx context.Context, mpesaConfirmation, pretiumStatus string, id int) error
}

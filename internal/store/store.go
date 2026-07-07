package store

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store interface {
	Pool() *pgxpool.Pool

	InsertNonCustodialLink(ctx context.Context, tx pgx.Tx, publicKey, phoneNumber string) error
	GetNonCustodialLinkByPublicKey(ctx context.Context, tx pgx.Tx, publicKey string) (*NonCustodialLink, error)
	GetNonCustodialLinkByPhone(ctx context.Context, tx pgx.Tx, phoneNumber string) (*NonCustodialLink, error)
	GetNonCustodialLinksByPhone(ctx context.Context, tx pgx.Tx, phoneNumber string) ([]NonCustodialLink, error)
	DeactivateNonCustodialLink(ctx context.Context, tx pgx.Tx, phoneNumber string) error
	DeactivateNonCustodialLinkByPublicKey(ctx context.Context, tx pgx.Tx, publicKey string) error

	InsertOfframp(ctx context.Context, tx pgx.Tx, pretiumID, phoneNumber, amountUSD, amountKES, txHash, tokenAddress, walletAddress string) error
	GetOfframpByPretiumID(ctx context.Context, tx pgx.Tx, pretiumID string) (*Offramp, error)
	GetOfframpByTxHash(ctx context.Context, tx pgx.Tx, txHash string) (*Offramp, error)
	GetOfframpsByPhone(ctx context.Context, tx pgx.Tx, phoneNumber string) ([]Offramp, error)
	GetOfframpsByWalletAddress(ctx context.Context, tx pgx.Tx, walletAddress string) ([]Offramp, error)
	GetStaleOfframps(ctx context.Context, tx pgx.Tx) ([]Offramp, error)
	GetPendingOfframps(ctx context.Context, tx pgx.Tx) ([]Offramp, error)
	GetRecentOfframps(ctx context.Context, tx pgx.Tx) ([]Offramp, error)
	UpdateOfframpStatus(ctx context.Context, tx pgx.Tx, pretiumStatus, pretiumID string) error
	UpdateOfframpMpesaConfirmation(ctx context.Context, tx pgx.Tx, mpesaConfirmation, pretiumStatus, pretiumID string) error

	InsertOnramp(ctx context.Context, tx pgx.Tx, pretiumID, phoneNumber string, amountUSD, amountKES, txHash, tokenAddress, walletAddress string) error
	GetOnrampByPretiumID(ctx context.Context, tx pgx.Tx, pretiumID string) (*Onramp, error)
	GetOnrampByTxHash(ctx context.Context, tx pgx.Tx, txHash string) (*Onramp, error)
	GetOnrampsByPhone(ctx context.Context, tx pgx.Tx, phoneNumber string) ([]Onramp, error)
	GetOnrampsByWalletAddress(ctx context.Context, tx pgx.Tx, walletAddress string) ([]Onramp, error)
	GetStaleOnramps(ctx context.Context, tx pgx.Tx) ([]Onramp, error)
	GetPendingOnramps(ctx context.Context, tx pgx.Tx) ([]Onramp, error)
	GetRecentOnramps(ctx context.Context, tx pgx.Tx) ([]Onramp, error)
	UpdateOnrampStatus(ctx context.Context, tx pgx.Tx, pretiumStatus, pretiumID string) error
	UpdateOnrampMpesaConfirmation(ctx context.Context, tx pgx.Tx, mpesaConfirmation, pretiumStatus, pretiumID string) error
	UpdateOnrampTxHash(ctx context.Context, tx pgx.Tx, txHash, pretiumStatus, pretiumID string) error
}

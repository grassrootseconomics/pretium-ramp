package store

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/tern/v2/migrate"
	"github.com/knadh/goyesql/v2"
)

type (
	Queries struct {
		InsertNonCustodialLink      string `query:"insert-non-custodial-link"`
		GetNonCustodialLinkByPubKey string `query:"get-non-custodial-link-by-public-key"`
		GetNonCustodialLinkByPhone  string `query:"get-non-custodial-link-by-phone"`
		GetNonCustodialLinksByPhone string `query:"get-non-custodial-links-by-phone"`
		DeactivateNonCustodialLink  string `query:"deactivate-non-custodial-link"`

		InsertOfframp             string `query:"insert-offramp"`
		GetOfframpByPretiumID     string `query:"get-offramp-by-pretium-id"`
		GetOfframpByTxHash        string `query:"get-offramp-by-tx-hash"`
		GetOfframpByPhone         string `query:"get-offramp-by-phone"`
		GetOfframpByWalletAddress string `query:"get-offramp-by-wallet-address"`
		GetStaleOfframps          string `query:"get-stale-offramps"`
		GetRecentOfframps         string `query:"get-recent-offramps"`
		UpdateOfframpStatus       string `query:"update-offramp-status"`
		UpdateOfframpMpesaConfirm string `query:"update-offramp-mpesa-confirmation"`

		InsertOnramp             string `query:"insert-onramp"`
		GetOnrampByPretiumID     string `query:"get-onramp-by-pretium-id"`
		GetOnrampByTxHash        string `query:"get-onramp-by-tx-hash"`
		GetOnrampByPhone         string `query:"get-onramp-by-phone"`
		GetOnrampByWalletAddress string `query:"get-onramp-by-wallet-address"`
		GetStaleOnramps          string `query:"get-stale-onramps"`
		GetRecentOnramps         string `query:"get-recent-onramps"`
		UpdateOnrampStatus       string `query:"update-onramp-status"`
		UpdateOnrampMpesaConfirm string `query:"update-onramp-mpesa-confirmation"`
	}

	PgOpts struct {
		Logg                 *slog.Logger
		DSN                  string
		MigrationsFolderPath string
		QueriesFolderPath    string
	}

	Pg struct {
		logg    *slog.Logger
		db      *pgxpool.Pool
		queries *Queries
	}
)

func NewPgStore(o PgOpts) (Store, error) {
	parsedConfig, err := pgxpool.ParseConfig(o.DSN)
	if err != nil {
		return nil, err
	}

	dbPool, err := pgxpool.NewWithConfig(context.Background(), parsedConfig)
	if err != nil {
		return nil, err
	}

	queries, err := loadQueries(o.QueriesFolderPath)
	if err != nil {
		return nil, err
	}

	if err := runMigrations(dbPool, o.MigrationsFolderPath); err != nil {
		return nil, err
	}
	o.Logg.Info("migrations ran successfully")

	return &Pg{
		logg:    o.Logg,
		db:      dbPool,
		queries: queries,
	}, nil
}

func (s *Pg) Pool() *pgxpool.Pool {
	return s.db
}

func (s *Pg) InsertNonCustodialLink(ctx context.Context, tx pgx.Tx, publicKey, phoneNumber string) error {
	_, err := tx.Exec(ctx, s.queries.InsertNonCustodialLink, publicKey, phoneNumber)
	return err
}

func (s *Pg) GetNonCustodialLinkByPublicKey(ctx context.Context, tx pgx.Tx, publicKey string) (*NonCustodialLink, error) {
	var link NonCustodialLink
	if err := pgxscan.Get(ctx, tx, &link, s.queries.GetNonCustodialLinkByPubKey, publicKey); err != nil {
		return nil, err
	}
	return &link, nil
}

func (s *Pg) GetNonCustodialLinkByPhone(ctx context.Context, tx pgx.Tx, phoneNumber string) (*NonCustodialLink, error) {
	var link NonCustodialLink
	if err := pgxscan.Get(ctx, tx, &link, s.queries.GetNonCustodialLinkByPhone, phoneNumber); err != nil {
		return nil, err
	}
	return &link, nil
}

func (s *Pg) GetNonCustodialLinksByPhone(ctx context.Context, tx pgx.Tx, phoneNumber string) ([]NonCustodialLink, error) {
	var links []NonCustodialLink
	if err := pgxscan.Select(ctx, tx, &links, s.queries.GetNonCustodialLinksByPhone, phoneNumber); err != nil {
		return nil, err
	}
	return links, nil
}

func (s *Pg) DeactivateNonCustodialLink(ctx context.Context, tx pgx.Tx, phoneNumber string) error {
	_, err := tx.Exec(ctx, s.queries.DeactivateNonCustodialLink, phoneNumber)
	return err
}

func (s *Pg) InsertOfframp(ctx context.Context, tx pgx.Tx, pretiumID, phoneNumber string, amountUSD, amountKES, txHash, tokenAddress, walletAddress string) error {
	_, err := tx.Exec(ctx, s.queries.InsertOfframp, pretiumID, phoneNumber, amountUSD, amountKES, txHash, tokenAddress, strings.ToLower(walletAddress))
	return err
}

func (s *Pg) GetOfframpByPretiumID(ctx context.Context, tx pgx.Tx, pretiumID string) (*Offramp, error) {
	var offramp Offramp
	if err := pgxscan.Get(ctx, tx, &offramp, s.queries.GetOfframpByPretiumID, pretiumID); err != nil {
		return nil, err
	}
	return &offramp, nil
}

func (s *Pg) GetOfframpByTxHash(ctx context.Context, tx pgx.Tx, txHash string) (*Offramp, error) {
	var offramp Offramp
	if err := pgxscan.Get(ctx, tx, &offramp, s.queries.GetOfframpByTxHash, txHash); err != nil {
		return nil, err
	}
	return &offramp, nil
}

func (s *Pg) GetOfframpsByPhone(ctx context.Context, tx pgx.Tx, phoneNumber string) ([]Offramp, error) {
	var offramps []Offramp
	if err := pgxscan.Select(ctx, tx, &offramps, s.queries.GetOfframpByPhone, phoneNumber); err != nil {
		return nil, err
	}
	return offramps, nil
}

func (s *Pg) GetOfframpsByWalletAddress(ctx context.Context, tx pgx.Tx, walletAddress string) ([]Offramp, error) {
	var offramps []Offramp
	if err := pgxscan.Select(ctx, tx, &offramps, s.queries.GetOfframpByWalletAddress, strings.ToLower(walletAddress)); err != nil {
		return nil, err
	}
	return offramps, nil
}

func (s *Pg) GetStaleOfframps(ctx context.Context, tx pgx.Tx) ([]Offramp, error) {
	var offramps []Offramp
	if err := pgxscan.Select(ctx, tx, &offramps, s.queries.GetStaleOfframps); err != nil {
		return nil, err
	}
	return offramps, nil
}

func (s *Pg) GetRecentOfframps(ctx context.Context, tx pgx.Tx) ([]Offramp, error) {
	var offramps []Offramp
	if err := pgxscan.Select(ctx, tx, &offramps, s.queries.GetRecentOfframps); err != nil {
		return nil, err
	}
	return offramps, nil
}

func (s *Pg) UpdateOfframpStatus(ctx context.Context, tx pgx.Tx, pretiumStatus, pretiumID string) error {
	_, err := tx.Exec(ctx, s.queries.UpdateOfframpStatus, pretiumStatus, pretiumID)
	return err
}

func (s *Pg) UpdateOfframpMpesaConfirmation(ctx context.Context, tx pgx.Tx, mpesaConfirmation, pretiumStatus, pretiumID string) error {
	_, err := tx.Exec(ctx, s.queries.UpdateOfframpMpesaConfirm, mpesaConfirmation, pretiumStatus, pretiumID)
	return err
}

func (s *Pg) InsertOnramp(ctx context.Context, tx pgx.Tx, pretiumID, phoneNumber string, amountUSD, amountKES, txHash, tokenAddress, walletAddress string) error {
	_, err := tx.Exec(ctx, s.queries.InsertOnramp, pretiumID, phoneNumber, amountUSD, amountKES, txHash, tokenAddress, strings.ToLower(walletAddress))
	return err
}

func (s *Pg) GetOnrampByPretiumID(ctx context.Context, tx pgx.Tx, pretiumID string) (*Onramp, error) {
	var onramp Onramp
	if err := pgxscan.Get(ctx, tx, &onramp, s.queries.GetOnrampByPretiumID, pretiumID); err != nil {
		return nil, err
	}
	return &onramp, nil
}

func (s *Pg) GetOnrampByTxHash(ctx context.Context, tx pgx.Tx, txHash string) (*Onramp, error) {
	var onramp Onramp
	if err := pgxscan.Get(ctx, tx, &onramp, s.queries.GetOnrampByTxHash, txHash); err != nil {
		return nil, err
	}
	return &onramp, nil
}

func (s *Pg) GetOnrampsByPhone(ctx context.Context, tx pgx.Tx, phoneNumber string) ([]Onramp, error) {
	var onramps []Onramp
	if err := pgxscan.Select(ctx, tx, &onramps, s.queries.GetOnrampByPhone, phoneNumber); err != nil {
		return nil, err
	}
	return onramps, nil
}

func (s *Pg) GetOnrampsByWalletAddress(ctx context.Context, tx pgx.Tx, walletAddress string) ([]Onramp, error) {
	var onramps []Onramp
	if err := pgxscan.Select(ctx, tx, &onramps, s.queries.GetOnrampByWalletAddress, strings.ToLower(walletAddress)); err != nil {
		return nil, err
	}
	return onramps, nil
}

func (s *Pg) GetStaleOnramps(ctx context.Context, tx pgx.Tx) ([]Onramp, error) {
	var onramps []Onramp
	if err := pgxscan.Select(ctx, tx, &onramps, s.queries.GetStaleOnramps); err != nil {
		return nil, err
	}
	return onramps, nil
}

func (s *Pg) GetRecentOnramps(ctx context.Context, tx pgx.Tx) ([]Onramp, error) {
	var onramps []Onramp
	if err := pgxscan.Select(ctx, tx, &onramps, s.queries.GetRecentOnramps); err != nil {
		return nil, err
	}
	return onramps, nil
}

func (s *Pg) UpdateOnrampStatus(ctx context.Context, tx pgx.Tx, pretiumStatus, pretiumID string) error {
	_, err := tx.Exec(ctx, s.queries.UpdateOnrampStatus, pretiumStatus, pretiumID)
	return err
}

func (s *Pg) UpdateOnrampMpesaConfirmation(ctx context.Context, tx pgx.Tx, mpesaConfirmation, pretiumStatus, pretiumID string) error {
	_, err := tx.Exec(ctx, s.queries.UpdateOnrampMpesaConfirm, mpesaConfirmation, pretiumStatus, pretiumID)
	return err
}

func loadQueries(queriesPath string) (*Queries, error) {
	parsedQueries, err := goyesql.ParseFile(queriesPath)
	if err != nil {
		return nil, err
	}

	loadedQueries := &Queries{}

	if err := goyesql.ScanToStruct(loadedQueries, parsedQueries, nil); err != nil {
		return nil, fmt.Errorf("failed to scan queries %v", err)
	}

	return loadedQueries, nil
}

func runMigrations(dbPool *pgxpool.Pool, migrationsPath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	conn, err := dbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	migrator, err := migrate.NewMigrator(ctx, conn.Conn(), "schema_version")
	if err != nil {
		return err
	}

	if err := migrator.LoadMigrations(os.DirFS(migrationsPath)); err != nil {
		return err
	}

	if err := migrator.Migrate(ctx); err != nil {
		return err
	}

	return nil
}

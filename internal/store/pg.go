package store

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/tern/v2/migrate"
	"github.com/knadh/goyesql/v2"
)

type (
	Queries struct {
		InsertNonCustodialLink      string `query:"insert-non-custodial-link"`
		GetNonCustodialLinkByPubKey string `query:"get-non-custodial-link-by-public-key"`
		GetNonCustodialLinkByPhone  string `query:"get-non-custodial-link-by-phone"`
		DeactivateNonCustodialLink  string `query:"deactivate-non-custodial-link"`

		InsertOfframp             string `query:"insert-offramp"`
		GetOfframpByPretiumID     string `query:"get-offramp-by-pretium-id"`
		GetOfframpByTxHash        string `query:"get-offramp-by-tx-hash"`
		GetOfframpByPhone         string `query:"get-offramp-by-phone"`
		UpdateOfframpStatus       string `query:"update-offramp-status"`
		UpdateOfframpMpesaConfirm string `query:"update-offramp-mpesa-confirmation"`

		InsertOnramp             string `query:"insert-onramp"`
		GetOnrampByPretiumID     string `query:"get-onramp-by-pretium-id"`
		GetOnrampByTxHash        string `query:"get-onramp-by-tx-hash"`
		GetOnrampByPhone         string `query:"get-onramp-by-phone"`
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

func (s *Pg) InsertNonCustodialLink(ctx context.Context, publicKey, phoneNumber string) error {
	_, err := s.db.Exec(ctx, s.queries.InsertNonCustodialLink, publicKey, phoneNumber)
	return err
}

func (s *Pg) GetNonCustodialLinkByPublicKey(ctx context.Context, publicKey string) (*NonCustodialLink, error) {
	var link NonCustodialLink
	if err := pgxscan.Get(ctx, s.db, &link, s.queries.GetNonCustodialLinkByPubKey, publicKey); err != nil {
		return nil, err
	}
	return &link, nil
}

func (s *Pg) GetNonCustodialLinkByPhone(ctx context.Context, phoneNumber string) (*NonCustodialLink, error) {
	var link NonCustodialLink
	if err := pgxscan.Get(ctx, s.db, &link, s.queries.GetNonCustodialLinkByPhone, phoneNumber); err != nil {
		return nil, err
	}
	return &link, nil
}

func (s *Pg) DeactivateNonCustodialLink(ctx context.Context, phoneNumber string) error {
	_, err := s.db.Exec(ctx, s.queries.DeactivateNonCustodialLink, phoneNumber)
	return err
}

func (s *Pg) InsertOfframp(ctx context.Context, pretiumID, phoneNumber string, amountUSD, amountKES, txHash, tokenAddress string) (int, error) {
	var id int
	err := s.db.QueryRow(ctx, s.queries.InsertOfframp, pretiumID, phoneNumber, amountUSD, amountKES, txHash, tokenAddress).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *Pg) GetOfframpByPretiumID(ctx context.Context, pretiumID string) (*Offramp, error) {
	var offramp Offramp
	if err := pgxscan.Get(ctx, s.db, &offramp, s.queries.GetOfframpByPretiumID, pretiumID); err != nil {
		return nil, err
	}
	return &offramp, nil
}

func (s *Pg) GetOfframpByTxHash(ctx context.Context, txHash string) (*Offramp, error) {
	var offramp Offramp
	if err := pgxscan.Get(ctx, s.db, &offramp, s.queries.GetOfframpByTxHash, txHash); err != nil {
		return nil, err
	}
	return &offramp, nil
}

func (s *Pg) GetOfframpsByPhone(ctx context.Context, phoneNumber string) ([]Offramp, error) {
	var offramps []Offramp
	if err := pgxscan.Select(ctx, s.db, &offramps, s.queries.GetOfframpByPhone, phoneNumber); err != nil {
		return nil, err
	}
	return offramps, nil
}

func (s *Pg) UpdateOfframpStatus(ctx context.Context, pretiumStatus, pretiumID string) error {
	_, err := s.db.Exec(ctx, s.queries.UpdateOfframpStatus, pretiumStatus, pretiumID)
	return err
}

func (s *Pg) UpdateOfframpMpesaConfirmation(ctx context.Context, mpesaConfirmation, pretiumStatus string, id int) error {
	_, err := s.db.Exec(ctx, s.queries.UpdateOfframpMpesaConfirm, mpesaConfirmation, pretiumStatus, id)
	return err
}

func (s *Pg) InsertOnramp(ctx context.Context, pretiumID, phoneNumber string, amountUSD, amountKES, txHash, tokenAddress string) (int, error) {
	var id int
	err := s.db.QueryRow(ctx, s.queries.InsertOnramp, pretiumID, phoneNumber, amountUSD, amountKES, txHash, tokenAddress).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *Pg) GetOnrampByPretiumID(ctx context.Context, pretiumID string) (*Onramp, error) {
	var onramp Onramp
	if err := pgxscan.Get(ctx, s.db, &onramp, s.queries.GetOnrampByPretiumID, pretiumID); err != nil {
		return nil, err
	}
	return &onramp, nil
}

func (s *Pg) GetOnrampByTxHash(ctx context.Context, txHash string) (*Onramp, error) {
	var onramp Onramp
	if err := pgxscan.Get(ctx, s.db, &onramp, s.queries.GetOnrampByTxHash, txHash); err != nil {
		return nil, err
	}
	return &onramp, nil
}

func (s *Pg) GetOnrampsByPhone(ctx context.Context, phoneNumber string) ([]Onramp, error) {
	var onramps []Onramp
	if err := pgxscan.Select(ctx, s.db, &onramps, s.queries.GetOnrampByPhone, phoneNumber); err != nil {
		return nil, err
	}
	return onramps, nil
}

func (s *Pg) UpdateOnrampStatus(ctx context.Context, pretiumStatus, pretiumID string) error {
	_, err := s.db.Exec(ctx, s.queries.UpdateOnrampStatus, pretiumStatus, pretiumID)
	return err
}

func (s *Pg) UpdateOnrampMpesaConfirmation(ctx context.Context, mpesaConfirmation, pretiumStatus string, id int) error {
	_, err := s.db.Exec(ctx, s.queries.UpdateOnrampMpesaConfirm, mpesaConfirmation, pretiumStatus, id)
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

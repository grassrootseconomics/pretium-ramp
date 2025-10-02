package store

import "github.com/jackc/pgx/v5/pgxpool"

type Store interface {
	Pool() *pgxpool.Pool
}

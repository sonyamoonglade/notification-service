package postgres

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Postgres struct {
	Pool *pgxpool.Pool
}

func New(ctx context.Context, dbURL string) (*Postgres, error) {
	connStr, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, err
	}
	pool, err := pgxpool.ConnectConfig(ctx, connStr)
	return &Postgres{Pool: pool}, nil
}

func (p *Postgres) CloseConn() {
	p.Pool.Close()
}

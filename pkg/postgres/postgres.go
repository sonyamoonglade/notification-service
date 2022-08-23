package postgres

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
)

type Postgres struct {
	Pool *pgxpool.Pool
}

func New(logger *zap.SugaredLogger, ctx context.Context, dbURL string) (*Postgres, error) {
	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, err
	}
	config.MinConns = 2
	config.MaxConns = 4

	pool, err := pgxpool.ConnectConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	return &Postgres{Pool: pool}, nil
}

func (p *Postgres) CloseConn() {
	p.Pool.Close()
}

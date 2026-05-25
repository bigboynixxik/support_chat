package postgress

import (
	"context"
	"fmt"
	"log/slog"
	"support_chat/pkg/logger"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultMaxConns        = 10
	defaultMaxConnLifeTime = time.Hour
	defaultMaxConnIdleTime = 30 * time.Minute
	defaultConnectTimeout  = 5 * time.Second
)

type Options struct {
	MaxConns        int32
	MaxConnLifeTime time.Duration
	MaxConnIdleTime time.Duration
	ConnectTimeout  time.Duration
}

func NewPool(ctx context.Context, dsn string, opts ...Options) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("postgress.Newpool, failed to parse dsn: %w", err)
	}
	opt := defaultOptions()
	if len(opts) > 0 {
		applyOptions(cfg, opts[0])
	} else {
		applyOptions(cfg, opt)
	}

	connectCtx, cancel := context.WithTimeout(ctx, cfg.PingTimeout)
	defer cancel()
	connectCtx = logger.WithContext(connectCtx, slog.Default())

	pool, err := pgxpool.NewWithConfig(connectCtx, cfg)
	if err != nil {
		return nil, fmt.Errorf("postgress.NewPool, failed to connect to postgres: %w", err)
	}
	if err := pool.Ping(connectCtx); err != nil {
		return nil, fmt.Errorf("postgress.NewPool, failed to ping postgres: %w", err)
	}
	return pool, nil
}

func MustNewPool(ctx context.Context, dsn string, opts ...Options) *pgxpool.Pool {
	pool, err := NewPool(ctx, dsn, opts...)
	if err != nil {
		panic(err)
	}
	return pool
}

func defaultOptions() Options {
	return Options{
		MaxConns:        defaultMaxConns,
		MaxConnLifeTime: defaultMaxConnLifeTime,
		MaxConnIdleTime: defaultMaxConnIdleTime,
		ConnectTimeout:  defaultConnectTimeout,
	}
}

func applyOptions(cfg *pgxpool.Config, opts Options) {
	if opts.MaxConns > 0 {
		cfg.MaxConns = opts.MaxConns
	}
	if opts.MaxConnLifeTime > 0 {
		cfg.MaxConnLifetime = opts.MaxConnLifeTime
	}
	if opts.MaxConnIdleTime > 0 {
		cfg.MaxConnIdleTime = opts.MaxConnIdleTime
	}
	if opts.ConnectTimeout > 0 {
		cfg.PingTimeout = opts.ConnectTimeout
	}
}

package gox

import (
	"context"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

const defaultDSN = "user=test password=test host=localhost port=5432 dbname=test sslmode=disable"

type ConnConfig struct {
	Host     string `env:"DB_HOST,required" envDefault:"localhost"`
	Database string `env:"DB_DATABASE,required" envDefault:"users"`
	Port     int    `env:"DB_PORT,required" envDefault:"5432"`
	User     string `env:"DB_USER,required" envDefault:"test"`
	Password string `env:"DB_PASSWORD,required" envDefault:"test"`
}

func NewPgPool(ctx context.Context, c *ConnConfig) (*pgxpool.Pool, error) {
	if c == nil {
		c = &ConnConfig{}
		if err := LoadFromEnv(c); err != nil {
			return nil, err
		}
	}

	conf, err := pgxpool.ParseConfig(defaultDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to parse default dsn %+v", err)
	}

	conf.ConnConfig.Host = c.Host
	conf.ConnConfig.Port = uint16(c.Port)
	conf.ConnConfig.Database = c.Database
	conf.ConnConfig.User = c.User
	conf.ConnConfig.Password = c.Password

	pool, err := pgxpool.ConnectConfig(ctx, conf)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return pool, nil
}

func IsNoRowError(err error) bool {
	return err == pgx.ErrNoRows
}

func IsDuplicateConstraintError(err error, constraintName string) bool {
	perr, ok := err.(*pgconn.PgError)
	return ok && perr.Code == "23505" && perr.ConstraintName == constraintName
}

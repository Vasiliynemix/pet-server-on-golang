package postgres

import (
	"PetProjectGo/internal/config"
	"PetProjectGo/pkg/logging"
	"embed"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
)

func NewPostgresConnection(log *logging.Logger, cfg *config.PostgresConnectionConfig) (*sqlx.DB, error) {
	const op = "postgres.NewPostgresConnection"

	db, err := sqlx.Open(config.DatabaseDriver, buildUri(cfg))
	if err != nil {
		log.Error("Postgres connection error", zap.String("op", op), zap.Error(err))
		return nil, err
	}

	db.DB.SetMaxIdleConns(cfg.Pool.MaxIdleConnections)
	db.DB.SetMaxOpenConns(cfg.Pool.MaxOpenConnections)
	db.DB.SetConnMaxIdleTime(cfg.Pool.IdleTimeout)

	return db, nil
}

//go:embed migrations/*.sql
var embedMigrations embed.FS

func Migrations(db *sqlx.DB, log *logging.Logger) error {
	const op = "postgres.Migrations"

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect(config.DatabaseDriver); err != nil {
		log.Error("Postgres migration error", zap.String("op", op), zap.Error(err))
		return err
	}

	if err := goose.Up(db.DB, "migrations"); err != nil {
		log.Error("Postgres migration error", zap.String("op", op), zap.Error(err))
		return err
	}

	log.Info("migrations success")

	return nil
}

func buildUri(cfg *config.PostgresConnectionConfig) string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Database,
	)
}

package database

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/spf13/viper"
)

func Connect() (*sqlx.DB, error) {
	dsn := viper.GetString("DB_DSN")
	return sqlx.Connect("pgx", dsn)
}

package models

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/paluras/product-recall-system/configs"
)

type DB struct {
	*sql.DB
}

func NewDB(config configs.DatabaseConfig) (*DB, error) {
	db, err := sql.Open("mysql", config.DSN())
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(config.MaxConns)
	db.SetMaxIdleConns(config.MaxIdle)
	db.SetConnMaxLifetime(config.Timeout)

	return &DB{db}, nil
}

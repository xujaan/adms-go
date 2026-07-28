package store

import (
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/go-sql-driver/mysql"
)

type Store struct {
	Adms *sqlx.DB
}

func New(dsn string) (*Store, error) {
	db, err := sqlx.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &Store{Adms: db}, nil
}

func (s *Store) Close() {
	if s.Adms != nil {
		s.Adms.Close()
	}
}

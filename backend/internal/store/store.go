package store

import (
	"database/sql"

	"myagents/internal/crypto"
)

type Store struct {
	db  *sql.DB
	box *crypto.Box
}

func New(db *sql.DB, box *crypto.Box) *Store {
	return &Store{db: db, box: box}
}

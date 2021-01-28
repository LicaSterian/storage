package handlers

import (
	"database/sql"

	"github.com/LicaSterian/storage/api/db"
)

// Handlers struct
type Handlers struct {
	db      *sql.DB
	storage db.Storage
}

// New constructor funcreceives a sql.DB pointer and returns a Handlers struct
func New(db *sql.DB, s db.Storage) Handlers {
	return Handlers{
		db:      db,
		storage: s,
	}
}

package repo

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type DB struct {
	conn *sql.DB
}

func NewDB(databaseURL string) (*DB, error) {
	conn, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &DB{conn: conn}

	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(5)

	return db, nil
}

func (d *DB) Close() error {
	return d.conn.Close()
}

func (d *DB) GetDB() *sql.DB {
	return d.conn
}

func (d *DB) BeginTx() (*sql.Tx, error) {
	return d.conn.Begin()
}

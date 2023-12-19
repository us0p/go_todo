package database

import (
	"database/sql"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

type DB interface{
    QueryRow(query string, args ...any) *sql.Row
    Query(query string, args ...any) (*sql.Rows, error)
    Exec(query string, args ...any) (sql.Result, error)
}

func OpenDatabase(rootFolder string) (*sql.DB, error) {
    return sql.Open("sqlite3", filepath.Join(rootFolder, "task.db"))
}

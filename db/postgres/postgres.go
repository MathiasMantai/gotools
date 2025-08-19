package postgres

import (
	"context"
	"fmt"
	"github.com/MathiasMantai/gotools/cli"
	// "github.com/MathiasMantai/gotools/db/util"
	_ "github.com/jackc/pgx/v5/stdlib"
	// "os"
	// "path/filepath"
	"database/sql"
)

type DbConnData struct {
	Server   string
	Port     string
	Database string
	User     string
	Pw       string
}

type PgSqlDb struct {
	DbObj    *sql.DB
	ConnData DbConnData
}

func (mdb *PgSqlDb) BeginTx(ctx context.Context, options *sql.TxOptions) (*sql.Tx, error) {
	return mdb.DbObj.BeginTx(ctx, options)
}

func (mdb *PgSqlDb) Exec(query string, args ...any) (sql.Result, error) {
	return mdb.DbObj.Exec(query, args...)
}

func (mdb *PgSqlDb) Query(query string, args ...any) (*sql.Rows, error) {
	return mdb.DbObj.Query(query, args...)
}

func (mdb *PgSqlDb) QueryRow(query string, args ...any) *sql.Row {
	return mdb.DbObj.QueryRow(query, args...)
}

// returns a PgSqlDb instance and an error
// if an error is returned the instance will be nil
func Connect(server string, port string, database string, user string, pw string) (*PgSqlDb, error) {
	var db PgSqlDb

	db.ConnData = DbConnData{
		Server:   server,
		Port:     port,
		Database: database,
		User:     user,
		Pw:       pw,
	}
	// urlExample := "postgres://username:password@localhost:5432/database_name"
	connectionString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, pw, server, port, database)
	conn, connError := sql.Open("pgx", connectionString)
	if connError != nil {
		return nil, connError
	}

	cli.PrintWithTimeAndColor("=> establishing Database connection with database "+database, "green", true)

	db.DbObj = conn

	return &db, nil
}

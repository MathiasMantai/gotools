package db

import (
	"context"
	"database/sql"
	"github.com/MathiasMantai/gotools/db/mssql"
	"github.com/MathiasMantai/gotools/db/mysql"
	"github.com/MathiasMantai/gotools/db/postgres"
	"github.com/MathiasMantai/gotools/db/sqlite"
)

type GotoolsDb interface {
	Query(string, ...interface{}) (*sql.Rows, error)
	QueryRow(string, ...interface{}) *sql.Row
	Exec(string, ...interface{}) (sql.Result, error)
	BeginTx(context.Context, *sql.TxOptions) (*sql.Tx, error)
}

type Db struct {
	DbObj GotoolsDb
}

type DbConnectOptions struct {
	Server string
	Port   string

	//is used as filepath for sqlite
	Database string

	User string
	Pw   string

	//mainly used for mysql. default is tcp
	Protocol string
}

func (d *Db) Connect(dbType string, options DbConnectOptions) error {

	var err error

	switch dbType {
	case "mysql":
		d.DbObj, err = mysql.Connect(
			options.Server,
			options.Port,
			options.Database,
			options.User,
			options.Pw,
			options.Protocol,
		)
	case "mssql":
		d.DbObj, err = mssql.Connect(
			options.Server,
			options.Port,
			options.Database,
			options.User,
			options.Pw,
		)
	case "sqlite":
		d.DbObj, err = sqlite.Connect(
			options.Database,
		)
	case "postgres":
		d.DbObj, err = postgres.Connect(
			options.Server,
			options.Port,
			options.Database,
			options.User,
			options.Pw,
		)

	}

	return err
}

func (mdb *Db) BeginTx(ctx context.Context, options *sql.TxOptions) (*sql.Tx, error) {
	return mdb.DbObj.BeginTx(ctx, options)
}

func (mdb *Db) Exec(query string, args ...interface{}) (sql.Result, error) {
	return mdb.DbObj.Exec(query, args...)
}

func (mdb *Db) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return mdb.DbObj.Query(query, args...)
}

func (mdb *Db) QueryRow(query string, args ...interface{}) *sql.Row {
	return mdb.DbObj.QueryRow(query, args...)
}

package mysql

import (
	"database/sql"
	"fmt"
	"github.com/MathiasMantai/gotools/cli"
	// "github.com/MathiasMantai/gotools/db"
	"github.com/go-sql-driver/mysql"
	"net"
	// "os"
	// "path/filepath"
	"context"
	"time"
)

type DbConnData struct {
	Server   string
	Port     string
	Database string
	User     string
	Pw       string
	Protocol string
}

type MySqlDb struct {
	DbObj    *sql.DB
	ConnData DbConnData
}

func (mdb *MySqlDb) BeginTx(ctx context.Context, options *sql.TxOptions) (*sql.Tx, error) {
	return mdb.DbObj.BeginTx(ctx, options)
}

func (mdb *MySqlDb) Exec(query string, args ...interface{}) (sql.Result, error) {
	return mdb.DbObj.Exec(query, args...)
}

func (mdb *MySqlDb) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return mdb.DbObj.Query(query, args...)
}

func (mdb *MySqlDb) QueryRow(query string, args ...interface{}) *sql.Row {
	return mdb.DbObj.QueryRow(query, args...)
}

func Connect(server string, port string, database string, user string, pw string, protocol string) (*MySqlDb, error) {

	var (
		cdb MySqlDb
	)

	cdb.ConnData = DbConnData{
		Server:   server,
		Port:     port,
		Database: database,
		User:     user,
		Pw:       pw,
		Protocol: protocol,
	}
	cfg := mysql.Config{
		User:                 user,
		Passwd:               pw,
		Net:                  protocol,
		Addr:                 net.JoinHostPort(server, port),
		DBName:               database,
		AllowNativePasswords: true,
		ParseTime:            true,
	}

	if cfg.Net == "" {
		cfg.Net = "tcp"
	}

	dsn := cfg.FormatDSN()
	fmt.Println(dsn)
	conn, connError := sql.Open("mysql", dsn)
	if connError != nil {
		return nil, connError
	}
	cli.PrintWithTimeAndColor("=> successfully connected to database "+database, "green", true)

	conn.SetConnMaxLifetime(time.Minute * 3)
	conn.SetMaxOpenConns(10)
	conn.SetMaxIdleConns(10)
	cdb.DbObj = conn
	return &cdb, nil
}

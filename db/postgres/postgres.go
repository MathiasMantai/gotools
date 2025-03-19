package postgres

import (
	"context"

	"fmt"
	"github.com/MathiasMantai/gotools/cli"
	"github.com/MathiasMantai/gotools/db"
	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"path/filepath"
)

type DbConnData struct {
	Server   string
	Port     string
	Database string
	User     string
	Pw       string
}

type PgSqlDb struct {
	DbObj    *pgx.Conn
	ConnData DbConnData
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
	conn, connError := pgx.Connect(context.Background(), connectionString)
	if connError != nil {
		return nil, connError
	}

	cli.PrintWithTimeAndColor("=> establishing Database connection with database "+database, "green", true)

	db.DbObj = conn

	return &db, nil
}

func (pg *PgSqlDb) Migrate(migrationPath string) error {

	sqlFiles, readDirError := os.ReadDir(migrationPath)
	if readDirError != nil {
		fmt.Println("Error reading dir")
		return readDirError
	}

	tx, txError := pg.DbObj.Begin(context.Background())
	if txError != nil {

		return txError
	}

	for _, sqlFile := range sqlFiles {
		fmt.Printf("=> executing migration %s\n", sqlFile.Name())
		queryFilePath := filepath.Join(migrationPath, sqlFile.Name())
		query, readFileError := os.ReadFile(queryFilePath)
		if readFileError != nil {
			return readFileError
		}

		_, queryError := pg.DbObj.Exec(context.Background(), string(query))
		if queryError != nil {
			tx.Rollback(context.Background())
			return queryError
		}

		fmt.Printf("=> migration %s executed successfully\n", db.RemoveFileExtension(sqlFile.Name()))
	}
	tx.Commit(context.Background())

	return nil
}

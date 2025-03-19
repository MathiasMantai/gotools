package mssql

import (
	"database/sql"
	"fmt"
	"github.com/MathiasMantai/gotools/cli"
	"github.com/MathiasMantai/gotools/db"
	_ "github.com/denisenkom/go-mssqldb"
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

type MssqlDb struct {
	DbObj    *sql.DB
	ConnData DbConnData
}

func Connect(server string, port string, database string, user string, pw string) (*MssqlDb, error) {
	var db MssqlDb

	db.ConnData = DbConnData{
		Server:   server,
		Port:     port,
		Database: database,
		User:     user,
		Pw:       pw,
	}

	connectionString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s;database=%s", db.ConnData.Server, db.ConnData.User, db.ConnData.Pw, db.ConnData.Port, db.ConnData.Database)
	cli.PrintWithTimeAndColor("=> establishing database connection with database "+database, "green", true)
	dbObj, ConnError := sql.Open("mssql", connectionString)
	if ConnError != nil {
		return nil, ConnError
	}

	db.DbObj = dbObj

	return &db, nil
}

func (ms *MssqlDb) Migrate(migrationPath string) error {

	sqlFiles, readDirError := os.ReadDir(migrationPath)
	if readDirError != nil {
		fmt.Println("Error reading dir")
		return readDirError
	}

	tx, txError := ms.DbObj.Begin()
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

		_, queryError := ms.DbObj.Exec(string(query))
		if queryError != nil {
			tx.Rollback()
			return queryError
		}

		fmt.Printf("=> migration %s executed successfully\n", db.RemoveFileExtension(sqlFile.Name()))
	}
	tx.Commit()

	return nil
}

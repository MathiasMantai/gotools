package mysql


import (
	"database/sql"
	"fmt"
		"github.com/MathiasMantai/gotools/db"
	"github.com/MathiasMantai/gotools/cli"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"path/filepath"
	"strings"
)

type DbConnData struct {
	Server   string
	Port     string
	Database string
	User     string
	Pw       string
}

/* MySQL */

type MySqlDb struct {
	DbObj    *sql.DB
	ConnData DbConnData
}

func Connect(server string, port string, database string, user string, pw string, protocol string) (*MySqlDb, error) {
	var (
		cdb              MySqlDb
		connectionString string
	)
	cdb.ConnData = DbConnData{
		Server:   server,
		Port:     port,
		Database: database,
		User:     user,
		Pw:       pw,
	}

	if strings.TrimSpace(protocol) == "" {
		connectionString = fmt.Sprintf("%s:%s@%s/%s", user, pw, server, database)
	} else {
		connectionString = fmt.Sprintf("%s:%s@%s(%s)/%s", user, pw, protocol, server, database)
	}
	fmt.Println(connectionString)
	conn, connError := sql.Open("mysql", connectionString)
	if connError != nil {
		return nil, connError
	}

	cdb.DbObj = conn
	return &cdb, nil
}

func (my *MySqlDb) MakeMigrations(migrationPath string) error {
	cli.PrintColor("=> Attempting to migrate database tables...", "blue", true)
	sqlFiles, readDirError := os.ReadDir(migrationPath)
	if readDirError != nil {
		cli.PrintColor("Error reading dir", "red", true)
		return readDirError
	}

	tx, txError := my.DbObj.Begin()
	if txError != nil {
		cli.PrintColor("x> Error starting transaction: "+txError.Error(), "red", true)
		return txError
	}

	for _, sqlFile := range sqlFiles {
		name := db.RemoveFileExtension(sqlFile.Name())
		cli.PrintColor(fmt.Sprintf("=> executing migration %s", name), "blue", true)
		queryFilePath := filepath.Join(migrationPath, sqlFile.Name())
		query, readFileError := os.ReadFile(queryFilePath)
		if readFileError != nil {
			cli.PrintColor("x> Error reading SQL file: "+readFileError.Error(), "red", true)
			return readFileError
		}

		_, queryError := my.DbObj.Exec(string(query))
		if queryError != nil {
			tx.Rollback()
			cli.PrintColor("x> SQL Error: "+queryError.Error(), "red", true)
			return queryError
		}

		cli.PrintColor(fmt.Sprintf("=> migration %s executed successfully", name), "green", true)
	}
	tx.Commit()

	return nil
}

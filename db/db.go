package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5"
	_ "github.com/mattn/go-sqlite3"
	"github.com/MathiasMantai/gotools/cli"
	"log"
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

type SqliteDb struct {
	DbObj    *sql.DB
	FilePath string
}

func ConnectSqlite(filePath string) (*SqliteDb, error) {
	var db SqliteDb

	db.FilePath = filePath

	fmt.Println("Starting database connection")
	dbObj, ConnError := sql.Open("sqlite3", db.FilePath)
	if ConnError != nil {
		return nil, ConnError
	}

	db.DbObj = dbObj

	return &db, nil
}

func (s *SqliteDb) Migrate(migrationDir string) error {
	dir, readDirError := os.ReadDir(migrationDir)
	if readDirError != nil {
		log.Fatal("=> Migration")
	}

	for _, fileName := range dir {
		file, err := os.ReadFile(filepath.Join(migrationDir, fileName.Name()))
		if err != nil {
			return err
		}

		_, queryError := s.DbObj.Exec(string(file))
		if queryError != nil {
			return queryError
		}
	}

	return nil
}

// run migration from an embedded directory
// reads every migrationfile separately and executes all qureries
func (s *SqliteDb) MigrateEmbedded(migrationDir embed.FS, dirName string) error {
	fmt.Println("=> Running migrations...")
	dir, readDirError := migrationDir.ReadDir("migrations")
	if readDirError != nil {
		log.Fatal("=> Migration dir could not be found")
	}

	for _, fileName := range dir {
		fmt.Printf("=> Running migration: %v\n", fileName.Name())
		var filePath string = dirName + "/" + fileName.Name()
		file, err := migrationDir.ReadFile(filePath)
		if err != nil {
			fmt.Printf("x> Error executing migration %v: %v\n", fileName.Name(), err.Error())
			return err
		}

		fmt.Println("=> " + string(file))

		_, queryError := s.DbObj.Exec(string(file))
		if queryError != nil {
			return queryError
		}
	}

	return nil
}

/* MSSQL */

type MssqlDb struct {
	DbObj    *sql.DB
	ConnData DbConnData
}

func ConnectMssql(server string, port string, database string, user string, pw string) (*MssqlDb, error) {
	var db MssqlDb

	db.ConnData = DbConnData{
		Server:   server,
		Port:     port,
		Database: database,
		User:     user,
		Pw:       pw,
	}

	connectionString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s;database=%s", db.ConnData.Server, db.ConnData.User, db.ConnData.Pw, db.ConnData.Port, db.ConnData.Database)

	fmt.Println("=> Establishing database connection")
	dbObj, ConnError := sql.Open("mssql", connectionString)
	if ConnError != nil {
		return nil, ConnError
	}

	db.DbObj = dbObj

	return &db, nil
}

func (ms *MssqlDb) MakeMigrations(migrationPath string) error {

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

		fmt.Printf("=> migration %s executed successfully\n", RemoveFileExtension(sqlFile.Name()))
	}
	tx.Commit()

	return nil
}

/* POSTGRES */
type PgSqlDb struct {
	DbObj    *pgx.Conn
	ConnData DbConnData
}

// returns a PgSqlDb instance and an error
// if an error is returned the instance will be nil
func ConnectPgSqlDb(server string, port string, database string, user string, pw string) (*PgSqlDb, error) {
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

	fmt.Println("=> Establishing database connection")

	db.DbObj = conn

	return &db, nil
}

func (pg *PgSqlDb) MakeMigrations(migrationPath string) error {

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

		fmt.Printf("=> migration %s executed successfully\n", RemoveFileExtension(sqlFile.Name()))
	}
	tx.Commit(context.Background())

	return nil
}

/* MySQL */

type MySqlDb struct {
	DbObj    *sql.DB
	ConnData DbConnData
}

func ConnectMySqlDb(server string, port string, database string, user string, pw string, protocol string) (*MySqlDb, error) {
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
		cli.PrintColor(fmt.Sprintf("=> executing migration %s", RemoveFileExtension(sqlFile.Name())), "blue", true)
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

		cli.PrintColor(fmt.Sprintf("=> migration %s executed successfully", RemoveFileExtension(sqlFile.Name())), "green", true)
	}
	tx.Commit()

	return nil
}
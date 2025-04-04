package sqlite

import (
	"database/sql"
	"embed"
	"fmt"
	"github.com/MathiasMantai/gotools/cli"
	_ "github.com/mattn/go-sqlite3"
	"log"
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

type SqliteDb struct {
	DbObj    *sql.DB
	FilePath string
}

func Connect(filePath string) (*SqliteDb, error) {
	var db SqliteDb

	db.FilePath = filePath

	cli.PrintWithTimeAndColor("=> establishing Database connection with database at path "+filePath, "green", true)
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

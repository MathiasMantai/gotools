package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/MathiasMantai/gotools/db/mssql"
	"github.com/MathiasMantai/gotools/db/mysql"
	"github.com/MathiasMantai/gotools/db/postgres"
	"github.com/MathiasMantai/gotools/db/sqlite"
)

type GotoolsDb interface {
	Query(string, ...any) (*sql.Rows, error)
	QueryRow(string, ...any) *sql.Row
	Exec(string, ...any) (sql.Result, error)
	BeginTx(context.Context, *sql.TxOptions) (*sql.Tx, error)
}

type Db struct {
	DbObj  GotoolsDb
	DbType string
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
	d.DbType = dbType
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

func (mdb *Db) Exec(query string, args ...any) (sql.Result, error) {
	return mdb.DbObj.Exec(query, args...)
}

func (mdb *Db) Query(query string, args ...any) (*sql.Rows, error) {
	return mdb.DbObj.Query(query, args...)
}

func (mdb *Db) QueryRow(query string, args ...any) *sql.Row {
	return mdb.DbObj.QueryRow(query, args...)
}

/*****************
	MIGRATIONS
******************/

type Migration struct {
	TableName   string
	Description string
	Fields      []MigrationField
	ForeignKeys []ForeignKey
}

type MigrationField struct {
	Name          string
	DataType      string
	Nullable      bool
	PrimaryKey    bool
	AutoIncrement bool
}

type ForeignKey struct {
	Name            string
	Column          string
	ReferenceTable  string
	ReferenceColumn string
}

type MigrationRunner interface {
	Run() error
	IsMigrationApplied(string) (bool, error)
	SetupMigrationTable() error
	// AddMigration(string, []MigrationField)
	LogMigration(string, string) error
}

func CreateMigrations(db *Db, migrations []Migration) error {
	switch db.DbType {
	case "mssql":
		{
			if mssqlDb, ok := db.DbObj.(*mssql.MssqlDb); ok {
				runner := mssql.CreateMigrationRunner(mssqlDb)
				runner.Migrations = []mssql.Migration{}

				//add migrations
				for _, migration := range migrations {
					var realMigration mssql.Migration

					realMigration.TableName = migration.TableName
					realMigration.Description = migration.Description
					realMigration.Fields = []mssql.MigrationField{}
					realMigration.ForeignKeys = []mssql.ForeignKey{}

					// fields
					for _, field := range migration.Fields {
						var realField mssql.MigrationField
						realField.Name = field.Name
						realField.DataType = field.DataType
						realField.Nullable = field.Nullable
						realField.PrimaryKey = field.PrimaryKey
						realField.AutoIncrement = field.AutoIncrement
						realMigration.Fields = append(realMigration.Fields, realField)
					}

					//foreign keys
					for _, fKey := range migration.ForeignKeys {
						var realFkey mssql.ForeignKey
						realFkey.Name = fKey.Name
						realFkey.Column = fKey.Column
						realFkey.ReferenceTable = fKey.ReferenceTable
						realFkey.ReferenceColumn = fKey.ReferenceColumn
						realMigration.ForeignKeys = append(realMigration.ForeignKeys, realFkey)
					}
					runner.Migrations = append(runner.Migrations, realMigration)
				}

				err := runner.Run()
				if err != nil {
					return err
				}

				return nil
			} else {
				return errors.New("database type supported but connection to database not established")
			}
		}
	case "mysql":
		{
			if mysqlDb, ok := db.DbObj.(*mysql.MySqlDb); ok {
				runner := mysql.CreateMigrationRunner(mysqlDb)

				//add migrations
				for _, migration := range migrations {
					var realMigration mysql.Migration

					realMigration.TableName = migration.TableName
					realMigration.Description = migration.Description
					realMigration.Fields = []mysql.MigrationField{}
					realMigration.ForeignKeys = []mysql.ForeignKey{}

					// fields
					for _, field := range migration.Fields {
						var realField mysql.MigrationField
						realField.Name = field.Name
						realField.DataType = field.DataType
						realField.Nullable = field.Nullable
						realField.PrimaryKey = field.PrimaryKey
						realField.AutoIncrement = field.AutoIncrement
						realMigration.Fields = append(realMigration.Fields, realField)
					}

					//foreign keys
					for _, fKey := range migration.ForeignKeys {
						var realFkey mysql.ForeignKey
						realFkey.Name = fKey.Name
						realFkey.Column = fKey.Column
						realFkey.ReferenceTable = fKey.ReferenceTable
						realFkey.ReferenceColumn = fKey.ReferenceColumn
						realMigration.ForeignKeys = append(realMigration.ForeignKeys, realFkey)
					}

					runner.Migrations = append(runner.Migrations, realMigration)
				}

				err := runner.Run()
				if err != nil {
					return err
				}

				return nil
			} else {
				return errors.New("database type supported but connection to database not established")
			}
		}
	case "sqlite":
		{
			if sqliteDb, ok := db.DbObj.(*sqlite.SqliteDb); ok {
				runner := sqlite.MigrationRunner{}
				runner.Db = sqliteDb
				for _, migration := range migrations {
					var realMigration sqlite.Migration
					

					realMigration.TableName = migration.TableName
					realMigration.Description = migration.Description
					realMigration.Fields = []sqlite.MigrationField{}
					realMigration.ForeignKeys = []sqlite.ForeignKey{}

					// fields
					for _, field := range migration.Fields {
						var realField sqlite.MigrationField
						realField.Name = field.Name
						realField.DataType = field.DataType
						realField.Nullable = field.Nullable
						realField.PrimaryKey = field.PrimaryKey
						realField.AutoIncrement = field.AutoIncrement
						realMigration.Fields = append(realMigration.Fields, realField)
					}

					//foreign keys
					for _, fKey := range migration.ForeignKeys {
						var realFkey sqlite.ForeignKey
						realFkey.Name = fKey.Name
						realFkey.Column = fKey.Column
						realFkey.ReferenceTable = fKey.ReferenceTable
						realFkey.ReferenceColumn = fKey.ReferenceColumn
						realMigration.ForeignKeys = append(realMigration.ForeignKeys, realFkey)
					}

					runner.Migrations = append(runner.Migrations, realMigration)
				}

				err := runner.Run()
				if err != nil {
					return err
				}
				return nil
			} else {
				return errors.New("database type supported but connection to database not established")
			}
		}
	default:
		return fmt.Errorf("unsupported Database type %v", db.DbType)
	}
}

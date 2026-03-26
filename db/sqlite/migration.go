package sqlite

import (
	"database/sql"
	"fmt"
	"github.com/MathiasMantai/gotools/cli"
	"strings"
)

type MigrationRunner struct {
	Migrations []Migration
	Db         *SqliteDb
}

func (mr *MigrationRunner) Run() error {
	err := mr.SetupMigrationTable()
	if err != nil {
		return fmt.Errorf("error creating migrations table: %v", err.Error())
	}

	for key, migration := range mr.Migrations {
		migrationText := fmt.Sprintf("migration %d - %s", key, migration.TableName)
		cli.PrintWithTimeAndColor("=> attempting to apply "+migrationText, "blue", true)

		if len(migration.Fields) == 0 {
			cli.PrintWithTimeAndColor(fmt.Sprintf("=> skipping %v since no fields were declared for table", migrationText), "yellow", true)
			continue
		}

		applied, err := mr.IsMigrationLogged(migration.TableName)
		if err != nil {
			return fmt.Errorf("error while checking if migration is already logged: %v", err.Error())
		}

		if !applied {
			tx, err := mr.Db.DbObj.Begin()
			if err != nil {
				return fmt.Errorf("error starting transaction for migration %s: %v", migration.TableName, err.Error())
			}

			createQuery := migration.CreateQuery()
			fmt.Println(createQuery)
			_, err = tx.Exec(createQuery)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("error creating table %s: %v", migration.TableName, err.Error())
			}

			// Migration loggen
			err = mr.LogMigrationTx(tx, migration.TableName, migration.Description)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("error logging table %s: %v", migration.TableName, err.Error())
			}

			err = tx.Commit()
			if err != nil {
				return fmt.Errorf("error committing transaction for migration %s: %v", migration.TableName, err.Error())
			}

			cli.PrintWithTimeAndColor("=> migration successfully applied and logged", "green", true)
		} else {
			cli.PrintWithTimeAndColor(fmt.Sprintf("=> %v already applied. Skipping...", migrationText), "yellow", true)
		}
	}

	if len(mr.Migrations) == 0 {
		cli.PrintWithTimeAndColor("=> no migrations to apply", "yellow", true)
	}
	return nil
}

func (mr *MigrationRunner) IsMigrationLogged(tableName string) (bool, error) {
	query := `
		SELECT 
			COUNT(*) 
		FROM 
			_migrations
		WHERE
			name = ?
	`
	var cnt int8
	err := mr.Db.DbObj.QueryRow(query, tableName).Scan(&cnt)
	if err != nil {
		return false, err
	}
	return cnt > 0, nil
}

func (mr *MigrationRunner) SetupMigrationTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS _migrations (
			name VARCHAR(255) NOT NULL UNIQUE, -- UNIQUE hinzugefügt
			description TEXT NULL,
			applied_at DATETIME NOT NULL DEFAULT current_timestamp 
		);
	`
	_, err := mr.Db.DbObj.Exec(query)
	return err
}

func (mr *MigrationRunner) LogMigrationTx(tx *sql.Tx, tableName string, description string) error {
	query := `
		INSERT INTO _migrations
		(
			name, 
			description
		)
		VALUES
		(
			?,
			?
		)
	`
	_, err := tx.Exec(query, tableName, description)
	return err
}

func (mr *MigrationRunner) LogMigration(tableName string, description string) error {
	query := `
		INSERT INTO _migrations
			(name, description)
		VALUES
			(?, ?)
	`
	_, err := mr.Db.DbObj.Exec(query, tableName, description)
	return err
}

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

func (m *Migration) CreateForeignKeyQueries() []string {
	queries := []string{}
	for _, fk := range m.ForeignKeys {
		query := fmt.Sprintf(`ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY(%s) REFERENCES %s (%s)`,
			m.TableName,
			fk.Name,
			fk.Column,
			fk.ReferenceTable,
			fk.ReferenceColumn,
		)
		queries = append(queries, query)
	}
	return queries
}

func (m *Migration) CreateQuery() string {
	var fieldDefs []string

	for _, field := range m.Fields {
		fieldDef := fmt.Sprintf("%s %s", field.Name, strings.ToUpper(field.DataType))

		if field.Nullable {
			fieldDef += " NULL"
		} else {
			fieldDef += " NOT NULL"
		}

		if field.PrimaryKey {
			if field.AutoIncrement && strings.ToUpper(field.DataType) == "INTEGER" {
				fieldDef += " PRIMARY KEY AUTOINCREMENT"
			} else {
				fieldDef += " PRIMARY KEY"
			}
		}

		fieldDefs = append(fieldDefs, fieldDef)
	}

	for _, fk := range m.ForeignKeys {
		fkDef := fmt.Sprintf("FOREIGN KEY (%s) REFERENCES %s (%s)",
			fk.Column,
			fk.ReferenceTable,
			fk.ReferenceColumn,
		)
		fieldDefs = append(fieldDefs, fkDef)
	}

	fieldsString := strings.Join(fieldDefs, ",\n\t")

	return fmt.Sprintf(`
		CREATE TABLE %v (
			%v
		)
	`, m.TableName, fieldsString)
}

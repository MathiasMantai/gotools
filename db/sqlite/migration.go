package sqlite

import (
	"fmt"
	"strings"

	"github.com/MathiasMantai/gotools/cli"
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
		cli.PrintWithTimeAndColor("=> attempting to apply " + migrationText, "blue", true)

		applied, err := mr.IsMigrationApplied(migration.TableName)
		if err != nil {
			return fmt.Errorf("error while checking if migration is already applied: %v", err.Error())
		}

		if !applied {
			createQuery := migration.CreateQuery()
			_, err := mr.Db.DbObj.Exec(createQuery)
			if err != nil {
				return fmt.Errorf("error creating table: %v", err.Error())
			}

			err = mr.LogMigration(migration.TableName, migration.Description)
			if err != nil {
				return fmt.Errorf("error logging table: %v", err.Error())
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

func (mr *MigrationRunner) IsMigrationApplied(tableName string) (bool, error) {
	query := `
		SELECT 
			COUNT(*) 
		FROM 
			sqlite_master
		WHERE
			type = 'table'
		AND 
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
			name VARCHAR(255) NOT NULL,
			description TEXT NULL,
			applied_at DATETIME NOT NULL DEFAULT current_timestamp 
		);
	`

	_, err := mr.Db.DbObj.Exec(query)
	return err
}


func (mr *MigrationRunner) LogMigration(tableName string, description string) error {
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
		query := fmt.Sprintf(`ALTER TABLE %s WITH CHECK ADD CONSTRAINT %s FOREIGN KEY(%s) REFERENCES %s (%s)`, 
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

		//a primary key fields auto incremnts

		if field.PrimaryKey {
			fieldDef += " PRIMARY KEY"
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
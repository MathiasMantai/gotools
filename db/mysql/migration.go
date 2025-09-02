package mysql

import (
	"fmt"
	"github.com/MathiasMantai/gotools/cli"
	"strings"
)

/*
*************

	MIGRATION

**************
*/

// A single migration. Will be translated into a CREATE statement
type Migration struct {
	TableName   string
	Description string
	Fields      []MigrationField
	ForeignKeys []ForeignKey
}

func (m *Migration) CreateForeignKeyQueries() []string {
	var queries []string

	for _, foreignKey := range m.ForeignKeys {
		query := fmt.Sprintf(`
			ALTER TABLE %s WITH CHECK ADD CONSTRAINT %s FOREIGN KEY(%s) REFERENCES %s (%s)
			`,
			m.TableName,
			foreignKey.Name,
			foreignKey.Column,
			foreignKey.ReferenceTable,
			foreignKey.ReferenceColumn,
		)

		queries = append(queries, query)
	}

	return queries
}

func (m *Migration) CreateQuery() string {
	var fields []string
	var primaryKeyFields []string

	for _, field := range m.Fields {
		fieldDef := fmt.Sprintf("%s %s", field.Name, strings.ToUpper(field.DataType))

		if field.Nullable {
			fieldDef += " NULL"
		} else {
			fieldDef += " NOT NULL"
		}

		if field.AutoIncrement {
			fieldDef += " AUTO_INCREMENT"
		}

		fields = append(fields, fieldDef)

		if field.PrimaryKey {
			primaryKeyFields = append(primaryKeyFields, field.Name)
		}
	}

	fieldsString := strings.Join(fields, ", \n\t\t")

	if len(primaryKeyFields) > 0 {
		fieldsString += ", \n\t\t"
		for i, pk := range primaryKeyFields {
			fieldsString += fmt.Sprintf(" PRIMARY KEY(%s)", pk)

			if i < len(primaryKeyFields)-1 {
				fieldsString += ","
			}

			fieldsString += "\n\t\t"
		}
	}

	return fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %v (
			%v
		)
	`, m.TableName, fieldsString)
}

/*
*******************

	MIGRATIONFIELD

********************
*/
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

/*
*
MIGRATIONRUNNER
*/
type MigrationRunner struct {
	Migrations []Migration
	Db         *MySqlDb
}

func (mr *MigrationRunner) SetupMigrationTable() error {

	query := `
		CREATE TABLE IF NOT EXISTS _migrations (
			id INT NOT NULL AUTO_INCREMENT,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			applied_at DATETIME NOT NULL ON UPDATE CURRENT_TIMESTAMP() DEFAULT CURRENT_TIMESTAMP(),
			PRIMARY KEY(id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`

	_, err := mr.Db.DbObj.Exec(query)
	if err != nil {
		return fmt.Errorf("error creating migrations table: %v", err.Error())
	}

	return nil
}

func (mr *MigrationRunner) IsMigrationApplied(name string) (bool, error) {
	query := `
		SELECT COUNT(*) 
		FROM
			information_schema.tables
		WHERE
			table_name = ?
	`

	var cnt int
	err := mr.Db.DbObj.QueryRow(query, name).Scan(&cnt)
	if err != nil {
		return false, fmt.Errorf("x> error checking if table %v already exists: %v", name, err.Error())
	}

	return cnt > 0, nil
}

func (mr *MigrationRunner) LogMigration(tableName string, description string) error {
	cli.PrintWithTimeAndColor("=> logging migration...", "blue", true)

	insertQuery := fmt.Sprintf(`
        INSERT INTO _migrations (
            name,
            description,
            applied_at
        )
        VALUES (
            '%s',
            '%s',
            CURRENT_TIMESTAMP()
        )
    `, strings.ReplaceAll(tableName, "'", "''"), strings.ReplaceAll(description, "'", "''"))

	_, err := mr.Db.DbObj.Exec(insertQuery)
	if err != nil {
		fmt.Println("Fehler beim Einfügen des Migration-Eintrags:", err)
		return err
	}

	return nil
}

func (mr *MigrationRunner) Run() error {
	//setup the migrations table

	err := mr.SetupMigrationTable()
	if err != nil {
		return err
	}

	for key, migration := range mr.Migrations {
		migrationText := fmt.Sprintf("migration %d - %s", key, migration.TableName)
		cli.PrintWithTimeAndColor("=> attempting to apply "+migrationText, "blue", true)

		applied, err := mr.IsMigrationApplied(migration.TableName)
		if err != nil {
			cli.PrintWithTimeAndColor(fmt.Sprintf("x> error checking whether migration is applied : %v", err.Error()), "red", true)
			return err
		}

		if !applied {
			createQuery := migration.CreateQuery()

			_, err := mr.Db.DbObj.Exec(createQuery)
			if err != nil {
				return err
			}

			fkQueries := migration.CreateForeignKeyQueries()
			if len(fkQueries) > 0 {
				cli.PrintWithTimeAndColor(fmt.Sprintf("=> applying %d foreign key(s) for %s...", len(fkQueries), migration.TableName), "blue", true)
				for i, fkQuery := range fkQueries {
					_, err := mr.Db.DbObj.Exec(fkQuery)
					if err != nil {
						cli.PrintWithTimeAndColor(fmt.Sprintf("x> error executing foreign key %d for %v: %v", i+1, migrationText, err.Error), "red", true)
						return err
					}
				}
				cli.PrintWithTimeAndColor("=> forein keys successfully applied", "green", true)
			}

			err = mr.LogMigration(migration.TableName, migration.Description)
			if err != nil {
				return err
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

func CreateMigrationRunner(db *MySqlDb) MigrationRunner {
	return MigrationRunner{
		Db:         db,
		Migrations: []Migration{},
	}
}

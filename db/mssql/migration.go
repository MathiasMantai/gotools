package mssql

import (
	"fmt"
	"github.com/MathiasMantai/gotools/cli"
	"strings"
)

type MigrationRunner struct {
	Migrations []Migration
	Db         *MssqlDb
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

type Migration struct {
	TableName   string
	Description string
	Fields      []MigrationField
	ForeignKeys []ForeignKey
}

func (m *MigrationRunner) Run() error {
	err := m.SetupMigrationTable()
	if err != nil {
		return err
	}

	var schema string
	err = m.Db.DbObj.QueryRow(`SELECT SCHEMA_NAME()`).Scan(&schema)
	if err != nil {
		return fmt.Errorf("x> could not determine default database schema: %w", err)
	}
	cli.PrintWithTimeAndColor(fmt.Sprintf("=> working in schema: [%s]", schema), "cyan", true)

	for key, migration := range m.Migrations {
		migrationText := fmt.Sprintf("migration %d - %s", key, migration.TableName)
		cli.PrintWithTimeAndColor("=> attempting to apply "+migrationText, "blue", true)

		applied, err := m.IsMigrationApplied(migration.TableName)
		if err != nil {
			cli.PrintWithTimeAndColor("x> error checking whether migration is applied: "+err.Error(), "red", true)
			return err
		}

		if !applied {
			createQuery := migration.CreateQuery()
			_, err := m.Db.DbObj.Exec(createQuery)
			if err != nil {
				cli.PrintWithTimeAndColor(fmt.Sprintf("x> error executing create table for %v: %v", migrationText, err.Error()), "red", true)
				return err
			}
			cli.PrintWithTimeAndColor("=> table created or already exists.", "green", true)

			fkQueries := migration.CreateForeignKeyQueries(schema)
			if len(fkQueries) > 0 {
				cli.PrintWithTimeAndColor(fmt.Sprintf("=> applying %d foreign key(s) for %s...", len(fkQueries), migration.TableName), "blue", true)
				for i, fkQuery := range fkQueries {
					_, err := m.Db.DbObj.Exec(fkQuery)
					if err != nil {
						cli.PrintWithTimeAndColor(fmt.Sprintf("x> error executing foreign key %d for %v: %v", i+1, migrationText, err.Error()), "red", true)
						return err
					}
				}
				cli.PrintWithTimeAndColor("=> foreign keys successfully applied.", "green", true)
			}

			err = m.LogMigration(migration.TableName, migration.Description)
			if err != nil {
				return err
			}
			cli.PrintWithTimeAndColor("=> migration successfully applied and logged.", "green", true)

		} else {
			cli.PrintWithTimeAndColor("=> "+migrationText+" already applied. Skipping...", "yellow", true)
		}
	}

	return nil
}

func (m *Migration) CreateForeignKeyQueries(schema string) []string {
	var queries []string

	for _, fk := range m.ForeignKeys {
		query := fmt.Sprintf(`
            IF NOT EXISTS (SELECT * FROM sys.foreign_keys 
                           WHERE name = '%s' AND parent_object_id = OBJECT_ID('[%s].[%s]'))
            BEGIN
                ALTER TABLE [%s].[%s] WITH CHECK ADD CONSTRAINT [%s] FOREIGN KEY([%s])
                REFERENCES [%s].[%s] ([%s]);

                ALTER TABLE [%s].[%s] CHECK CONSTRAINT [%s];
            END
        `,
			fk.Name,
			schema, m.TableName,
			schema, m.TableName, fk.Name, fk.Column,
			schema, fk.ReferenceTable, fk.ReferenceColumn,
			schema, m.TableName, fk.Name)

		queries = append(queries, strings.TrimSpace(query))
	}

	return queries
}

func (ms *MigrationRunner) LogMigration(tableName string, description string) error {
	cli.PrintWithTimeAndColor("=> logging migration...", "blue", true)

	var schema string
	err := ms.Db.DbObj.QueryRow(`
        SELECT DEFAULT_SCHEMA_NAME 
        FROM sys.database_principals 
        WHERE name = CURRENT_USER
    `).Scan(&schema)

	if err != nil {
		fmt.Println("Fehler beim Ermitteln des Schemas:", err)
		return err
	}

	createTableQuery := fmt.Sprintf(`
        IF NOT EXISTS (SELECT * FROM INFORMATION_SCHEMA.TABLES 
                     WHERE TABLE_SCHEMA = '%s' AND TABLE_NAME = 'migrations')
        BEGIN
            CREATE TABLE %s.[migrations] (
                id INT PRIMARY KEY IDENTITY(1,1),
                name NVARCHAR(255) NOT NULL,
                description NVARCHAR(MAX),
                applied_at DATETIME NOT NULL
            )
        END
    `, schema, schema)

	_, err = ms.Db.DbObj.Exec(createTableQuery)
	if err != nil {
		fmt.Println("Fehler beim Erstellen der migrations-Tabelle:", err)
		return err
	}

	insertQuery := fmt.Sprintf(`
        INSERT INTO %s.[migrations] (
            name,
            description,
            applied_at
        )
        VALUES (
            '%s',
            '%s',
            GETDATE()
        )
    `, schema, strings.Replace(tableName, "'", "''", -1), strings.Replace(description, "'", "''", -1))

	_, err = ms.Db.DbObj.Exec(insertQuery)
	if err != nil {
		fmt.Println("Fehler beim Einfügen des Migration-Eintrags:", err)
		return err
	}

	return nil
}

func (ms *MigrationRunner) IsMigrationApplied(name string) (bool, error) {
	query := fmt.Sprintf(`
        DECLARE @MigrationName NVARCHAR(255) = '%s';
        DECLARE @CurrentSchema NVARCHAR(128);
        DECLARE @Count INT = 0;
        DECLARE @SQL NVARCHAR(MAX);
        
        SELECT @CurrentSchema = DEFAULT_SCHEMA_NAME 
        FROM sys.database_principals 
        WHERE name = CURRENT_USER;
        
        SET @SQL = N'SELECT @CountOUT = COUNT(*) FROM ' + QUOTENAME(@CurrentSchema) + '.migrations WHERE name = @MigrationName';
        
        EXEC sp_executesql @SQL, 
             N'@MigrationName NVARCHAR(255), @CountOUT INT OUTPUT', 
             @MigrationName = @MigrationName, 
             @CountOUT = @Count OUTPUT;
        
        SELECT @Count;
    `, name)

	var count int
	err := ms.Db.DbObj.QueryRow(query).Scan(&count)

	if err != nil {
		return false, fmt.Errorf("fehler beim Prüfen der Migration: %v", err)
	}
	return count > 0, nil
}

func (mr *MigrationRunner) SetupMigrationTable() error {
	query := `

	DECLARE @CurrentSchema NVARCHAR(128);
	SELECT @CurrentSchema = DEFAULT_SCHEMA_NAME 
	FROM sys.database_principals 
	WHERE name = CURRENT_USER;
	IF NOT EXISTS (SELECT * FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = @CurrentSchema AND TABLE_NAME = 'migrations')

	BEGIN
		DECLARE @SQL NVARCHAR(MAX) = 'CREATE TABLE ' + QUOTENAME(@CurrentSchema) + '.[migrations] (
			id INT PRIMARY KEY IDENTITY(1,1),
			name NVARCHAR(255) NOT NULL,
			description NVARCHAR(MAX),
			applied_at DATETIME NOT NULL
		)';

		EXEC sp_executesql @SQL;
	END
	`

	_, err := mr.Db.DbObj.Exec(query)
	if err != nil {
		return fmt.Errorf("x> error creating migrationstable: %v", err)
	}
	return nil
}

func (mr *MigrationRunner) ConvertToStruct(targetDir string, index int, jsonMapping bool) string {
	var targetMigration Migration = mr.Migrations[index]

	fields := ""
	for _, field := range targetMigration.Fields {
		fields += fmt.Sprintf("%v %v\n", field.Name, field.DataType)

	}

	return fmt.Sprintf("type %v struct {%v}", targetMigration.TableName, fields)
}

func CreateMigrationRunner(db *MssqlDb) MigrationRunner {
	return MigrationRunner{
		Db:         db,
		Migrations: []Migration{},
	}
}

func (mr *MigrationRunner) AddMigration(tableName string, fields []MigrationField) {
	mr.Migrations = append(mr.Migrations, Migration{
		TableName: tableName,
		Fields:    fields,
	})
}

func (m *Migration) CreateQuery() string {
	var fields []string
	var primaryKeyFields []string

	for _, field := range m.Fields {
		fieldDef := fmt.Sprintf("[%s] %s", field.Name, strings.ToUpper(field.DataType))

		if field.AutoIncrement {
			fieldDef += " IDENTITY(1, 1)"
		}

		if field.Nullable {
			fieldDef += " NULL"
		} else {
			fieldDef += " NOT NULL"
		}

		fields = append(fields, fieldDef)

		if field.PrimaryKey {
			primaryKeyFields = append(primaryKeyFields, fmt.Sprintf("[%s]", field.Name))
		}
	}

	fieldsString := strings.Join(fields, ",\n\t\t")

	if len(primaryKeyFields) > 0 {
		pkConstraintName := fmt.Sprintf("PK_%s", m.TableName)
		pkFieldsString := strings.Join(primaryKeyFields, ", ")
		fieldsString += fmt.Sprintf(",\n\t\tCONSTRAINT [%s] PRIMARY KEY CLUSTERED (%s)", pkConstraintName, pkFieldsString)
	}

	query := fmt.Sprintf(`
		DECLARE @CurrentSchema NVARCHAR(128);
		SELECT @CurrentSchema = DEFAULT_SCHEMA_NAME 
		FROM sys.database_principals 
		WHERE name = CURRENT_USER;

		IF NOT EXISTS (SELECT * FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = @CurrentSchema AND TABLE_NAME = '%s')
		BEGIN
				DECLARE @SQL NVARCHAR(MAX) = N'CREATE TABLE ' + QUOTENAME(@CurrentSchema) + '.[%s] (
					%s
				)';

				EXEC sp_executesql @SQL;
		END
	`, m.TableName, m.TableName, fieldsString)

	return strings.TrimSpace(query)
}

func (m *Migration) AddField(name string, dataType string) {
	m.Fields = append(m.Fields, MigrationField{
		Name:     name,
		DataType: dataType,
	})
}

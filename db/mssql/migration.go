package mssql


import (
	"fmt"
	"github.com/MathiasMantai/gotools/cli"
	"strings"
)

type MigrationRunner struct {
	Migrations           []Migration
	Db                   *MssqlDb
}


func (m *MigrationRunner) Run() error {

	err := m.SetupMigrationTable()
	if err != nil {
		return err
	}

	for key, migration := range m.Migrations {

		migrationText := fmt.Sprintf("migration %d - %v", key, migration.TableName)
		cli.PrintWithTimeAndColor("=> attempting to apply "+migrationText, "blue", true)
		//check if migration already exists
		applied, err := m.IsMigrationApplied(migration.TableName)
		if err != nil {
			cli.PrintWithTimeAndColor("x> error checking whether migration is applied: "+err.Error(), "red", true)
			return err
		}
		// fmt.Printf("migration applied: %f\n", applied)
		if !applied {
			query := migration.CreateQuery()
			// fmt.Println(query)

			_, err := m.Db.DbObj.Exec(query)
			if err != nil {
				cli.PrintWithTimeAndColor(fmt.Sprintf("x> error executing %v: %v", migrationText, err.Error()), "red", true)
				return err
			}
			cli.PrintWithTimeAndColor("=> migration successfully applied", "green", true)
			err = m.LogMigration(migration.TableName, migration.Description)
			if err != nil {
				return err
			}
		} else {
			cli.PrintWithTimeAndColor("=> "+migrationText+" already applied. Skipping...", "yellow", true)
		}
	}

	return nil
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

type Migration struct {
	TableName   string
	Description string
	Fields      []MigrationField
}

func (m *Migration) CreateQuery() string {
	var fields string
	for index, field := range m.Fields {
		fields += fmt.Sprintf("%v %v", field.Name, strings.ToUpper(field.DataType))
		if field.AutoIncrement {
			fields += " IDENTITY(1, 1)"
		}
		if index < len(m.Fields)-1 {
			fields += ","
		}
		fields += "\n\t"
	}

	//create query and check if table already exists
	query := fmt.Sprintf(`

		DECLARE @CurrentSchema NVARCHAR(128);
		SELECT @CurrentSchema = DEFAULT_SCHEMA_NAME 
		FROM sys.database_principals 
		WHERE name = CURRENT_USER;

		IF NOT EXISTS (SELECT * FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = @CurrentSchema AND  TABLE_NAME = '%v')
		BEGIN
				DECLARE @SQL NVARCHAR(MAX) = 'CREATE TABLE ' + QUOTENAME(@CurrentSchema) + '.[%v] (
					%v
				)';

				EXEC sp_executesql @SQL;
		END
	`, m.TableName, m.TableName, fields)

	return strings.TrimSpace(query)
}

func (m *Migration) AddField(name string, dataType string) {
	m.Fields = append(m.Fields, MigrationField{
		Name:     name,
		DataType: dataType,
	})
}

// defines a single table field of a
type MigrationField struct {
	//name of field
	Name string
	//datatype of field
	DataType string
	//if the field is a primary key
	AutoIncrement bool
}

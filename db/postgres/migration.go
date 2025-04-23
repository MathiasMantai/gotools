package postgres

import (
	"context"
	"fmt"
	"github.com/MathiasMantai/gotools/cli"
	"path/filepath"
	"strings"
)

type MigrationRunner struct {
	Migrations []Migration
	Db         *PgSqlDb
}

func CreateMigrationRunner(db *PgSqlDb) MigrationRunner {
	return MigrationRunner{
		Db:         db,
		Migrations: []Migration{},
	}
}

func (mr *MigrationRunner) AddMigration(tableName string, description string, fields []MigrationField) {
	mr.Migrations = append(mr.Migrations, Migration{
		TableName:   tableName,
		Description: description,
		Fields:      fields,
	})
}

func (mr *MigrationRunner) Run() error {
	ctx := context.Background()

	err := mr.SetupMigrationTable(ctx)
	if err != nil {
		return fmt.Errorf("failed to setup migration table: %w", err)
	}

	for i, migration := range mr.Migrations {
		migrationText := fmt.Sprintf("migration %d - %s", i, migration.TableName)
		cli.PrintWithTimeAndColor("=> attempting to apply "+migrationText, "blue", true)

		applied, err := mr.IsMigrationApplied(ctx, migration.TableName)
		if err != nil {
			cli.PrintWithTimeAndColor(fmt.Sprintf("x> error checking whether migration '%s' is applied: %v", migration.TableName, err), "red", true)
			return fmt.Errorf("checking migration '%s' failed: %w", migration.TableName, err)
		}

		if !applied {
			query := migration.CreateQuery()

			_, err := mr.Db.DbObj.Exec(ctx, query)
			if err != nil {
				cli.PrintWithTimeAndColor(fmt.Sprintf("x> error executing %s: %v", migrationText, err), "red", true)
				return fmt.Errorf("executing migration '%s' failed: %w", migration.TableName, err)
			}

			cli.PrintWithTimeAndColor("=> "+migrationText+" successfully applied", "green", true)
			err = mr.LogMigration(ctx, migration.TableName, migration.Description)
			if err != nil {
				cli.PrintWithTimeAndColor(fmt.Sprintf("x> error logging %s: %v", migrationText, err), "red", true)
				return fmt.Errorf("logging migration '%s' failed: %w", migration.TableName, err)
			}
		} else {
			cli.PrintWithTimeAndColor("=> "+migrationText+" already applied. Skipping...", "yellow", true)
		}
	}

	cli.PrintWithTimeAndColor("=> All migrations processed.", "green", true)
	return nil
}

func (mr *MigrationRunner) LogMigration(ctx context.Context, tableName string, description string) error {
	cli.PrintWithTimeAndColor("=> logging migration '"+tableName+"'...", "blue", true)

	insertQuery := `
        INSERT INTO migrations (name, description, applied_at)
        VALUES ($1, $2, NOW())
    `

	_, err := mr.Db.DbObj.Exec(ctx, insertQuery, tableName, description)
	if err != nil {
		return fmt.Errorf("inserting migration log for '%s' failed: %w", tableName, err)
	}

	cli.PrintWithTimeAndColor("=> logged migration '"+tableName+"' successfully", "green", true)
	return nil
}

func (mr *MigrationRunner) IsMigrationApplied(ctx context.Context, name string) (bool, error) {

	query := `
		SELECT COUNT(*)
		FROM migrations
		WHERE name = $1
	`

	var count int

	err := mr.Db.DbObj.QueryRow(ctx, query, name).Scan(&count)

	if err != nil {
		return false, fmt.Errorf("querying migration status for '%s' failed: %w", name, err)
	}

	return count > 0, nil
}

func (mr *MigrationRunner) SetupMigrationTable(ctx context.Context) error {
	query := `
	CREATE TABLE IF NOT EXISTS migrations (
		id SERIAL PRIMARY KEY,                -- Auto-inkrementierender Primärschlüssel
		name VARCHAR(255) NOT NULL UNIQUE,    -- Name der Migration, sollte eindeutig sein
		description TEXT,                     -- Beschreibung der Migration
		applied_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW() -- Zeitstempel der Anwendung
	);
	`
	cli.PrintWithTimeAndColor("=> ensuring migrations table exists...", "blue", true)
	_, err := mr.Db.DbObj.Exec(ctx, query)
	if err != nil {
		cli.PrintWithTimeAndColor(fmt.Sprintf("x> error creating/checking migrations table: %v", err), "red", true)
		return fmt.Errorf("creating/checking migrations table failed: %w", err)
	}
	cli.PrintWithTimeAndColor("=> migrations table check complete.", "green", true)
	return nil
}

func (mr *MigrationRunner) ConvertToStruct(targetDir string, index int, jsonMapping bool) string {
	if index < 0 || index >= len(mr.Migrations) {
		return "// Error: Invalid migration index"
	}
	targetMigration := mr.Migrations[index]

	fields := ""
	structName := ToGoStructName(targetMigration.TableName)

	for _, field := range targetMigration.Fields {
		goType := MapPgTypeToGo(field.DataType)
		goName := ToGoFieldName(field.Name)
		jsonTag := ""
		if jsonMapping {
			jsonTag = fmt.Sprintf("`json:\"%s\"`", field.Name)
		}
		fields += fmt.Sprintf("\t%s %s %s\n", goName, goType, jsonTag)
	}

	return fmt.Sprintf("package %s // Oder ein passender Paketname\n\nimport \"time\" // Beispielimport\n\ntype %s struct {\n%s}", filepath.Base(targetDir), structName, fields)
}

func ToGoStructName(dbName string) string {
	parts := strings.Split(dbName, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, "")
}

func ToGoFieldName(dbName string) string {
	return ToGoStructName(dbName)
}

func MapPgTypeToGo(pgType string) string {
	pgType = strings.ToLower(pgType)

	switch {
	case strings.HasPrefix(pgType, "int"), strings.HasPrefix(pgType, "serial"), strings.HasPrefix(pgType, "smallint"):
		return "int"
	case strings.HasPrefix(pgType, "bigint"), strings.HasPrefix(pgType, "bigserial"):
		return "int64"
	case strings.HasPrefix(pgType, "numeric"), strings.HasPrefix(pgType, "decimal"):
		return "float64"
	case strings.HasPrefix(pgType, "real"), strings.HasPrefix(pgType, "double precision"):
		return "float64"
	case strings.HasPrefix(pgType, "text"), strings.HasPrefix(pgType, "varchar"), strings.HasPrefix(pgType, "char"), strings.HasPrefix(pgType, "uuid"):
		return "string"
	case strings.HasPrefix(pgType, "timestamp"):
		return "time.Time"
	case strings.HasPrefix(pgType, "date"):
		return "time.Time"
	case strings.HasPrefix(pgType, "bool"):
		return "bool"
	case strings.HasPrefix(pgType, "bytea"):
		return "[]byte"
	case strings.HasPrefix(pgType, "json"), strings.HasPrefix(pgType, "jsonb"):
		return "[]byte"
	default:
		return "interface{}"
	}
}

type Migration struct {
	TableName   string
	Description string
	Fields      []MigrationField
}

func (m *Migration) CreateQuery() string {
	var fieldDefs []string
	primaryKeyDefined := false

	for _, field := range m.Fields {
		fieldDef := fmt.Sprintf("%q %s", field.Name, strings.ToUpper(field.DataType))

		if field.AutoIncrement && !primaryKeyDefined {
			fieldDef = fmt.Sprintf("%q SERIAL PRIMARY KEY", field.Name)
			primaryKeyDefined = true
		} else if field.AutoIncrement {
			fmt.Printf("WARNUNG: Mehrere AutoIncrement-Felder in Tabelle '%s' definiert. Nur das erste wird als SERIAL PRIMARY KEY behandelt.\n", m.TableName)
		}
		fieldDefs = append(fieldDefs, fieldDef)
	}

	fieldsSQL := strings.Join(fieldDefs, ",\n\t")

	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %q (
			%s
		);
	`, m.TableName, fieldsSQL)

	return strings.TrimSpace(query)
}

func (m *Migration) AddField(name string, dataType string, autoIncrement bool) {
	m.Fields = append(m.Fields, MigrationField{
		Name:          name,
		DataType:      dataType,
		AutoIncrement: autoIncrement,
	})
}

type MigrationField struct {
	Name          string
	DataType      string
	AutoIncrement bool
}

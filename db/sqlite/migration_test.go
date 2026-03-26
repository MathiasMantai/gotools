package sqlite

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"strings"
	"testing"
)

func getTestDb(t *testing.T) *SqliteDb {
	db, err := sql.Open("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}
	return &SqliteDb{DbObj: db}
}

func closeTestDb(t *testing.T, db *SqliteDb) {
	if db.DbObj != nil {
		err := db.DbObj.Close()
		if err != nil {
			t.Errorf("Failed to close database: %v", err)
		}
	}
}

func TestSetupMigrationTable(t *testing.T) {
	db := getTestDb(t)
	defer closeTestDb(t, db)

	runner := &MigrationRunner{Db: db}
	err := runner.SetupMigrationTable()
	if err != nil {
		t.Fatalf("SetupMigrationTable failed: %v", err)
	}

	var tableName string
	err = db.DbObj.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='_migrations'").Scan(&tableName)
	if err != nil {
		t.Fatalf("Checking _migrations table existence failed: %v", err)
	}
	if tableName != "_migrations" {
		t.Fatalf("Expected table _migrations, got %s", tableName)
	}

	var sqlDef string
	err = db.DbObj.QueryRow("SELECT sql FROM sqlite_master WHERE type='table' AND name='_migrations'").Scan(&sqlDef)
	if err != nil {
		t.Fatalf("Failed to get SQL definition for _migrations: %v", err)
	}
	if !strings.Contains(sqlDef, "name VARCHAR(255) NOT NULL UNIQUE") {
		t.Errorf("Expected _migrations table to have UNIQUE constraint on 'name', got: %s", sqlDef)
	}
}

func TestIsMigrationLogged(t *testing.T) {
	db := getTestDb(t)
	defer closeTestDb(t, db)

	runner := &MigrationRunner{Db: db}
	runner.SetupMigrationTable()

	logged, err := runner.IsMigrationLogged("test_table")
	if err != nil {
		t.Fatalf("IsMigrationLogged failed: %v", err)
	}
	if logged {
		t.Error("Expected migration 'test_table' not to be logged, but it was")
	}

	_, err = db.DbObj.Exec("INSERT INTO _migrations (name, description) VALUES (?, ?)", "test_table", "Test Description")
	if err != nil {
		t.Fatalf("Failed to insert test migration: %v", err)
	}

	logged, err = runner.IsMigrationLogged("test_table")
	if err != nil {
		t.Fatalf("IsMigrationLogged failed: %v", err)
	}
	if !logged {
		t.Error("Expected migration 'test_table' to be logged, but it was not")
	}
}

func TestRunNoMigrations(t *testing.T) {
	db := getTestDb(t)
	defer closeTestDb(t, db)

	runner := &MigrationRunner{Db: db, Migrations: []Migration{}}
	err := runner.Run()
	if err != nil {
		t.Fatalf("Run with no migrations failed: %v", err)
	}
}

func TestRunSingleMigrationSuccess(t *testing.T) {
	db := getTestDb(t)
	defer closeTestDb(t, db)

	migrations := []Migration{
		{
			TableName:   "users",
			Description: "Creates users table",
			Fields: []MigrationField{
				{Name: "id", DataType: "INTEGER", PrimaryKey: true, AutoIncrement: true},
				{Name: "name", DataType: "TEXT", Nullable: false},
				{Name: "email", DataType: "TEXT", Nullable: false},
			},
		},
	}

	runner := &MigrationRunner{Db: db, Migrations: migrations}
	err := runner.Run()
	if err != nil {
		t.Fatalf("Run single migration failed: %v", err)
	}

	var tableName string
	err = db.DbObj.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='users'").Scan(&tableName)
	if err != nil {
		t.Fatalf("Checking 'users' table existence failed: %v", err)
	}
	if tableName != "users" {
		t.Fatalf("Expected table 'users', got %s", tableName)
	}

	logged, err := runner.IsMigrationLogged("users")
	if err != nil {
		t.Fatalf("IsMigrationLogged failed after run: %v", err)
	}
	if !logged {
		t.Error("Expected migration 'users' to be logged, but it was not")
	}

	rows, err := db.DbObj.Query("PRAGMA table_info(users)")
	if err != nil {
		t.Fatalf("Failed to get table info for 'users': %v", err)
	}
	defer rows.Close()

	var cols []string
	for rows.Next() {
		var cid int
		var name string
		var typ string
		var notnull int
		var dfltValue sql.NullString
		var pk int
		err = rows.Scan(&cid, &name, &typ, &notnull, &dfltValue, &pk)
		if err != nil {
			t.Fatalf("Failed to scan table info row: %v", err)
		}
		cols = append(cols, fmt.Sprintf("%s %s PK:%d NOTNULL:%d", name, typ, pk, notnull))
	}

	expectedCols := []string{
		"id INTEGER PK:1 NOTNULL:1",
		"name TEXT PK:0 NOTNULL:1",
		"email TEXT PK:0 NOTNULL:1",
	}

	if len(cols) != len(expectedCols) {
		t.Errorf("Expected %d columns, got %d", len(expectedCols), len(cols))
	}
	for i, col := range cols {
		if i < len(expectedCols) && col != expectedCols[i] {
			t.Errorf("Column mismatch at index %d. Expected '%s', got '%s'", i, expectedCols[i], col)
		}
	}
}

func TestRunMigrationAlreadyLogged(t *testing.T) {
	db := getTestDb(t)
	defer closeTestDb(t, db)

	runner := &MigrationRunner{Db: db}
	runner.SetupMigrationTable()

	_, err := db.DbObj.Exec("INSERT INTO _migrations (name, description) VALUES (?, ?)", "existing_table", "An existing table")
	if err != nil {
		t.Fatalf("Failed to log existing migration: %v", err)
	}

	migrations := []Migration{
		{
			TableName:   "existing_table",
			Description: "Should be skipped",
			Fields: []MigrationField{
				{Name: "id", DataType: "INTEGER", PrimaryKey: true},
			},
		},
		{
			TableName:   "new_table",
			Description: "Should be applied",
			Fields: []MigrationField{
				{Name: "id", DataType: "INTEGER", PrimaryKey: true},
			},
		},
	}

	runner.Migrations = migrations
	err = runner.Run()
	if err != nil {
		t.Fatalf("Run with existing migration failed: %v", err)
	}

	var newTableName string
	err = db.DbObj.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='new_table'").Scan(&newTableName)
	if err != sql.ErrNoRows && err != nil {
		t.Fatalf("Checking 'new_table' existence failed: %v", err)
	}
	if newTableName != "new_table" {
		t.Errorf("Expected table 'new_table', but it was not created or name was wrong")
	}

	var count int
	err = db.DbObj.QueryRow("SELECT COUNT(*) FROM _migrations WHERE name = 'existing_table'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count existing_table in _migrations: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 entry for 'existing_table' in _migrations, got %d", count)
	}
}

func TestRunMigrationWithForeignKeys(t *testing.T) {
	db := getTestDb(t)
	defer closeTestDb(t, db)

	usersMigration := Migration{
		TableName:   "users_fk",
		Description: "Creates users_fk table for foreign key tests",
		Fields: []MigrationField{
			{Name: "id", DataType: "INTEGER", PrimaryKey: true, AutoIncrement: true},
			{Name: "name", DataType: "TEXT", Nullable: false},
		},
	}

	postsMigration := Migration{
		TableName:   "posts_fk",
		Description: "Creates posts_fk table with foreign key to users_fk",
		Fields: []MigrationField{
			{Name: "id", DataType: "INTEGER", PrimaryKey: true, AutoIncrement: true},
			{Name: "title", DataType: "TEXT", Nullable: false},
			{Name: "user_id", DataType: "INTEGER", Nullable: false},
		},
		ForeignKeys: []ForeignKey{
			{Name: "fk_user", Column: "user_id", ReferenceTable: "users_fk", ReferenceColumn: "id"},
		},
	}

	runner := &MigrationRunner{Db: db, Migrations: []Migration{usersMigration, postsMigration}}
	err := runner.Run()
	if err != nil {
		t.Fatalf("Run with foreign key migrations failed: %v", err)
	}

	var usersTableName, postsTableName string
	db.DbObj.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='users_fk'").Scan(&usersTableName)
	db.DbObj.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='posts_fk'").Scan(&postsTableName)

	if usersTableName != "users_fk" || postsTableName != "posts_fk" {
		t.Errorf("Expected 'users_fk' and 'posts_fk' tables to be created, got '%s' and '%s'", usersTableName, postsTableName)
	}

	rows, err := db.DbObj.Query("PRAGMA foreign_key_list(posts_fk)")
	if err != nil {
		t.Fatalf("Failed to get foreign key list for 'posts_fk': %v", err)
	}
	defer rows.Close()

	foundFk := false
	for rows.Next() {
		var id, seq int
		var table, from, to, onUpdate, onDelete, match string
		err := rows.Scan(&id, &seq, &table, &from, &to, &onUpdate, &onDelete, &match)
		if err != nil {
			t.Fatalf("Failed to scan foreign key row: %v", err)
		}
		if table == "users_fk" && from == "user_id" && to == "id" {
			foundFk = true
			break
		}
	}
	if !foundFk {
		t.Error("Expected foreign key 'user_id' referencing 'users_fk(id)' not found on 'posts_fk'")
	}
}

func TestCreateQueryAutoIncrement(t *testing.T) {
	migration := Migration{
		TableName:   "test_autoinc",
		Description: "Test auto increment",
		Fields: []MigrationField{
			{Name: "id", DataType: "INTEGER", PrimaryKey: true, AutoIncrement: true},
			{Name: "value", DataType: "TEXT", Nullable: false},
		},
	}

	query := migration.CreateQuery()
	expectedSubString := "id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT"
	if !strings.Contains(query, expectedSubString) {
		t.Errorf("Expected query to contain '%s', but got:\n%s", expectedSubString, query)
	}

	//auto increment should be ignored for this test case
	migrationNonInteger := Migration{
		TableName: "test_non_integer_autoinc",
		Fields: []MigrationField{
			{Name: "id", DataType: "TEXT", PrimaryKey: true, AutoIncrement: true},
		},
	}
	queryNonInteger := migrationNonInteger.CreateQuery()
	expectedNonIntegerSubString := "id TEXT NOT NULL PRIMARY KEY"
	if !strings.Contains(queryNonInteger, expectedNonIntegerSubString) {
		t.Errorf("Expected query for non-integer PK to contain '%s', but got:\n%s", expectedNonIntegerSubString, queryNonInteger)
	}
	if strings.Contains(queryNonInteger, "AUTOINCREMENT") {
		t.Errorf("Expected query for non-integer PK NOT to contain 'AUTOINCREMENT', but it did:\n%s", queryNonInteger)
	}
}

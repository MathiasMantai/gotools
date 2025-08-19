package mysql

import (
	// "database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMysqlDb_Exec(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	mysqlDb := &MySqlDb{
		DbObj: db,
	}
	expectedQuery := `
		INSERT INTO 
			users
		(name, email)
		VALUES
		(?, ?)
	`
	
	mock.ExpectExec(regexp.QuoteMeta(expectedQuery)).WithArgs("Max Mustermann", "maxmustermann@test.com").WillReturnResult(sqlmock.NewResult(1,1))

	result, err := mysqlDb.Exec(expectedQuery, "Max Mustermann", "maxmustermann@test.com")
    require.NoError(t, err)
    assert.NotNil(t, result)
    
    rowsAffected, _ := result.RowsAffected()
    assert.Equal(t, int64(1), rowsAffected)
    
    assert.NoError(t, mock.ExpectationsWereMet())
}
package mssql

import (
	"fmt"
	// "errors"
)

// add functionality to create methods, both for normal db objects and transaction objects (tx)

type Scaffold struct {
	Options      ScaffoldOptions `json:"scaffold_options"`
	QueryBuilder *QueryBuilder   `json:"query_builder"`
}

type ScaffoldOptions struct {
	DatabaseCredentials DatabaseCredentials `json:"database_credentials"`
	ModelDirectory      string              `json:"model_directory"`
	ModelDirectoryName  string              `json:"model_directory_name"`
	TableFilter         TableFilter         `json:"table_filter"`
	Schema              string              `json:"schema"`
	
}

type TableFilter struct {
	Prefix        string `json:"prefix"`
	CaseSensitive bool   `json:"case_sensitive"`
}

type Table struct {
	TableName string   `json:"table_name"`
	Columns   []Column `json:"columns"`
}

type Column struct {
	ColumnName      string  `json:"column_name"`
	DataType        string  `json:"data_type"`
	MaxLength       *int32  `json:"max_length"`
	IsNullable      bool    `json:"is_nullable"`
	DefaultValue    *string `json:"default_value"`
	OrdinalPosition int16   `json:"ordinal_position"`
}



/**

Scan the database and analyse the structure of select/all tables (option to only select specific ones)

*/

func (sc *Scaffold) Run(db *MssqlDb) error {
	if sc.QueryBuilder == nil {
		sc.QueryBuilder = NewQueryBuilder()
	}

	tables, err := sc.ScanDb(db)

	if err != nil {
		return err
	}

	fmt.Println(tables)

	// err = sc.CreateScaffolding(tables)

	// if err != nil {
	// 	return err
	// }

	return nil
}

func (sc *Scaffold) ScanDb(db *MssqlDb) ([]Table, error) {

	// SELECT TABLE_NAME
	// FROM INFORMATION_SCHEMA.TABLES
	// WHERE TABLE_TYPE = 'BASE TABLE' AND TABLE_CATALOG='MeuserIT' AND TABLE_NAME LIKE 'rad_%'

	query := sc.QueryBuilder.Select([]string{
		"TABLE_NAME",
	}).
		From("INFORMATION_SCHEMA.TABLES").
		Where("TABLE_TYPE", "=", "BASE TABLE").
		And("TABLE_CATALOG", "=", sc.Options.DatabaseCredentials.Database)

	if sc.Options.TableFilter.Prefix != "" {
		query = query.And("TABLE_NAME", "LIKE", fmt.Sprintf("%v%%", sc.Options.TableFilter.Prefix))
	}

	rows, err := db.DbObj.Query(query.Get())

	if err != nil {
		return nil, err
	}

	var rs []Table

	for rows.Next() {
		var tableName string
		err = rows.Scan(&tableName)

		//prepare query for columns
		query = sc.QueryBuilder.Select([]string{
			"COLUMN_NAME",
			"DATA_TYPE",
			"CHARACTER_MAXIMUM_LENGTH",
			"IS_NULLABLE",
			"COLUMN_DEFAULT",
			"ORDINAL_POSITION",
		}).
			From("INFORMATION_SCHEMA.COLUMNS").
			Where("TABLE", "=", tableName).
			OrderBy([]string{
				"ORDINAL_POSITION",
			})

		if err != nil {
			return rs, err
		}

		var table Table
		table.TableName = tableName

		rowsC, err := db.DbObj.Query(query.Get())
		if err != nil {
			return nil, err
		}

		for rowsC.Next() {
			var column Column
			err = rowsC.Scan(
				&column.ColumnName,
				&column.DataType,
				&column.MaxLength,
				&column.IsNullable,
				&column.DefaultValue,
			)

			if err != nil {
				return nil, err
			}

			table.Columns = append(table.Columns, column)
		}

		rs = append(rs, table)
	}

	return rs, nil
}


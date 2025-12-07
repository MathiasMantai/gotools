package mssql

import (
	"errors"
	"fmt"
	"github.com/MathiasMantai/gotools/osutil"
	s "github.com/MathiasMantai/gotools/string"
	"os"
	"path"
	"path/filepath"
	"strings"
	"slices"
)

// add functionality to create methods, both for normal db objects and transaction objects (tx)

type Scaffold struct {
	Options      ScaffoldOptions `json:"scaffold_options"`
	QueryBuilder *QueryBuilder   `json:"query_builder"`
}

type ScaffoldOptions struct {
	DatabaseCredentials DatabaseCredentials `json:"database_credentials"`
	ModelDirectory      string              `json:"model_directory"`
	TableFilter         TableFilter         `json:"table_filter"`
	Schema              string              `json:"schema"`
	Format              Format              `json:"format"`
}

type Format struct {
	//when set to true, all struct names and struct parameters will be attempted to be converted to camel case
	UseCamelCase bool `json:"use_camel_case"`
	//only used in combination with UseCamelCase. Otherwise this option will be ignored
	UseUpperCamelCase bool `json:"use_upper_camel_case"`
	UseJson           bool `json:"use_json"`
	UseYaml           bool `json:"use_yaml"`
	PublicStruct      bool `json:"public_struct"`
	PublicMethods     bool `json:"public_methods"`
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
	IsNullable      string    `json:"is_nullable"`
	DefaultValue    *string `json:"default_value"`
	OrdinalPosition int16   `json:"ordinal_position"`
}

type StructAttribute struct {
	AttributeName string `json:"attribute_name"`
	AttributeType string `json:"attribute_type"`
	JsonDefinition string `json:"json_definition"`
}

func (c *Column) TranslateColumnDefinition(useCamelCase bool, upperCamelCase bool, useJson bool) (StructAttribute, error) {
	var (
		attributeName  string
		attributeType  string
		jsonDefinition string
	)
	if strings.TrimSpace(c.ColumnName) == "" {
		return StructAttribute{}, errors.New("column name is empty")
	}

	attributeName = c.ColumnName
	if useCamelCase {
		attributeName = s.SnakeCaseToCamelCase(attributeName, true)
	}

    // Mapping SQL Data Types to Go Data Types
    switch strings.ToLower(c.DataType) {
    case "int", "tinyint", "smallint", "bigint":
        attributeType = "int"
    case "float", "real":
        attributeType = "float32"
    case "decimal", "numeric", "money", "smallmoney":
        attributeType = "float64"
    case "bit":
        attributeType = "bool"
    case "varchar", "nvarchar", "text", "ntext", "char", "nchar":
        attributeType = "string"
    case "datetime", "datetime2", "date", "smalldatetime", "time":
        attributeType = "time.Time"
    case "uniqueidentifier":
        attributeType = "string"
    case "varbinary", "image":
        attributeType = "[]byte"
    default:
        attributeType = "interface{}"
	}


    // Handle nullability
    if c.IsNullable == "YES" {
        attributeType = "*" + attributeType
    }


	//json without uppercase is useless
    if useJson && upperCamelCase {
        jsonTag := s.SnakeCaseToCamelCase(c.ColumnName, false)
        jsonDefinition = fmt.Sprintf("`json:\"%v\"`", jsonTag)
    }

	// return fmt.Sprintf("\t%v %v %v", attributeName, attributeType, jsonDefinition), nil
	return StructAttribute{
		AttributeName: attributeName,
		AttributeType: attributeType,
		JsonDefinition: jsonDefinition,
	}, nil
}

/**

Scan the database and analyse the structure of select/all tables (option to only select specific ones)

*/

func (sc *Scaffold) Run(db *MssqlDb) error {
	if sc.QueryBuilder == nil {
		sc.QueryBuilder = NewQueryBuilder()
	}

	tables, err := sc.ScanDb(db)

	fmt.Println("before scan db error handling")

	if err != nil {
		return err
	}

	fmt.Println(tables)

	err = sc.CreateScaffolding(tables)

	if err != nil {
		return err
	}

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
		Where("TABLE_TYPE", "=", "'BASE TABLE'").
		And("TABLE_CATALOG", "=", fmt.Sprintf("'%v'", sc.Options.DatabaseCredentials.Database))

	if sc.Options.TableFilter.Prefix != "" {
		query = query.And("TABLE_NAME", "LIKE", fmt.Sprintf("'%v%%'", sc.Options.TableFilter.Prefix))
	}

	rows, err := db.DbObj.Query(query.Get())

	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var rs []Table

	for rows.Next() {
		var tableName string
		err = rows.Scan(&tableName)

		if err != nil {
			fmt.Println("scanerror")
			return rs, err
		}

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
			Where("TABLE_NAME", "=", fmt.Sprintf("'%v'", tableName)).
			OrderBy([]string{
				"ORDINAL_POSITION",
			})

		var table Table
		table.TableName = tableName



		rowsC, err := db.DbObj.Query(query.Get())
		if err != nil {
			fmt.Println(err.Error())
			return nil, err
		}
		defer rowsC.Close()
		cols, _ := rowsC.Columns()
		fmt.Println("Available columns:", cols)
		fmt.Println("before column fetch")
		fmt.Println(rowsC)
		for rowsC.Next() {
			fmt.Println("inside next")
			var column Column
			err = rowsC.Scan(
				&column.ColumnName,
				&column.DataType,
				&column.MaxLength,
				&column.IsNullable,
				&column.DefaultValue,
				&column.OrdinalPosition,
			)

			if err != nil {
				fmt.Println(err.Error())
				return nil, err
			}
			fmt.Println("inside column fetch")
			fmt.Println(column)

			table.Columns = append(table.Columns, column)
		}

		rs = append(rs, table)
	}

	return rs, nil
}



func (sc *Scaffold) CreateScaffolding(tables []Table) error {

	//check if model dir exists and create if not
	if !osutil.FileExists(sc.Options.ModelDirectory) {
		os.Mkdir(sc.Options.ModelDirectory, 0755)
	}

	for _, table := range tables {
		fileName := fmt.Sprintf("%v.go", table.TableName)

		//define attributes
		var attributes []StructAttribute
		for _, column := range table.Columns {
			attributeDefinition, err := column.TranslateColumnDefinition(
				sc.Options.Format.UseCamelCase,
				sc.Options.Format.UseUpperCamelCase,
				sc.Options.Format.UseJson,
			)

			if err != nil {
				return err
			}

			attributes = append(attributes, attributeDefinition)
		}

		//package name
		packageName := strings.ToLower(path.Base(sc.Options.ModelDirectory))

		//imports
		imports := []string{
			"\"github.com/MathiasMantai/gotools/db/mssql\"",
		}

		var attributeStrings []string

		for _, attributeDefinition := range attributes {
			attributeStrings = append(
				attributeStrings,
				fmt.Sprintf(
					"\t%v %v %v",
					attributeDefinition.AttributeName, 
					attributeDefinition.AttributeType,
					attributeDefinition.JsonDefinition,
				),
			)
		}

		for _, attribute := range attributeStrings {
			if strings.Contains(attribute, "time.Time") && !slices.Contains(imports, "\"time\"") {
				imports = append(imports, "\"time\"")
			}
		}

		if len(imports) == 1 {
			imports[0] = "\t" + imports[0]
		}

		importString := strings.Join(imports, "\n\t")

		structName := table.TableName
		if sc.Options.Format.UseCamelCase {
			structName = s.SnakeCaseToCamelCase(structName, sc.Options.Format.UseUpperCamelCase)
		}

		structTemplate := fmt.Sprintf(
			"type %v struct {\n%v\n}",
			structName,
			strings.Join(attributeStrings, "\n"),
		)

		//methods
		crudMethods := createCrudMethods(structName, attributes, table.Columns)

		fileContent := fmt.Sprintf(
			"package %v\n\nimport (\n%v\n)\n\n%v\n\n%v",
			packageName,
			importString,
			structTemplate,
			crudMethods,
		)

		targetFile := filepath.Join(sc.Options.ModelDirectory, fileName)
		err := os.WriteFile(targetFile, []byte(fileContent), 0755)
		if err != nil {
			return err
		}
	}

	return nil
}

func createCrudMethods(structName string, structAttributes []StructAttribute, columns []Column) string {

	structAbbreviation := strings.ToLower(structName[0:1])

	columnString := createColumnString(columns)
	insertValues := createInsertValues(structAttributes)

	createCode := []string{
		fmt.Sprintf("func (%v *%v) Create(db *mssql.MssqlDb) error {\n", structAbbreviation, structName),
		fmt.Sprint("\tquery := db.Qb.InsertInto("),
		fmt.Sprint("\t\t\"%v\","),
		fmt.Sprintf("\t\t[]string {"),
		columnString,
		"\t\t},",
		"\t\t[]string {",
		insertValues,
		"\t\t},",
		"\t)",
		"\t_, err := db.DbObj.Exec(query.Get())",
		"\treturn err",
		"}",
	}
	createCodeString := strings.Join(createCode, "\n")
	// createString := fmt.Sprintf(
	// 	"func (%v *%v) Create(db *mssql.MssqlDb) {\n%v\n}",
	// 	structAbbreviation,
	// 	structName,
	// 	createCodeString,
	// )

	readString := fmt.Sprintf("")

	updateString := fmt.Sprintf("")

	deleteString := fmt.Sprintf("")

	return fmt.Sprintf(
		"%v\n\n%v\n\n%v\n\n%v",
		createCodeString,
		readString,
		updateString,
		deleteString,
	)
}

type Test struct {
	Username string
	Password string
}

func createColumnString(columns []Column) string {
	var rs string
	for i, column := range columns {
		rs += fmt.Sprintf("\t\t\t\"%v\",", column.ColumnName)
		if i < len(columns) -1 {
			rs += "\n"
		}
	}

	return rs
}


func createInsertValues(structAttributes []StructAttribute) string {
	var rs string
	for i, attribute := range structAttributes {
		rs += fmt.Sprintf("\t\t\t\"%v\",", attribute.AttributeName)
		if i < len(structAttributes) -1 {
			rs += "\n"
		}
	}
	return rs
}
package mssql

import (
	"fmt"
	"strings"
)

type QueryBuilder struct {
	Query string
}

func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{}
}

// return query as a string
func (q *QueryBuilder) Get() string {
	res := strings.TrimSpace(q.Query)
	q.Query = ""
	return res
}

/* SELECT */
func (q *QueryBuilder) SelectOne(selectOptions string) *QueryBuilder {
	q.Query += fmt.Sprintf("SELECT %s", selectOptions)
	return q
}

func (q *QueryBuilder) Select(selectOptions []string) *QueryBuilder {
	q.Query += "SELECT "
	for i, selectOption := range selectOptions {
		q.Query += selectOption

		if i < len(selectOptions)-1 {
			q.Query += ", "
		} else {
			q.Query += " "
		}
	}
	return q
}

func (q *QueryBuilder) SelectAll() *QueryBuilder {
	q.Query += "SELECT * "
	return q
}

/* FROM */

func (q *QueryBuilder) From(table string) *QueryBuilder {
	q.Query += fmt.Sprintf("FROM %s ", table)
	return q
}

/* WHERE */
func (q *QueryBuilder) Where(column string, comparisonOperator string, value string) *QueryBuilder {
	q.Query += fmt.Sprintf("WHERE %s %s %s ", column, comparisonOperator, value)
	return q
}

/* AND + OR */
func (q *QueryBuilder) And(column string, comparisonOperator string, value string) *QueryBuilder {
	q.Query += fmt.Sprintf("AND %s %s %s ", column, comparisonOperator, value)
	return q
}

func (q *QueryBuilder) Or(column string, comparisonOperator string, value string) *QueryBuilder {
	q.Query += fmt.Sprintf("OR %s %s %s ", column, comparisonOperator, value)
	return q
}

func (q *QueryBuilder) OrderBy(columns []string) *QueryBuilder {
	q.Query += "ORDER BY "
	length := len(columns)
	for key, column := range columns {
		q.Query += fmt.Sprintf("%v", column)
		if key < length {
			q.Query += ", "
		}
	}

	return q
}

func (q *QueryBuilder) GroupBy(columns []string) *QueryBuilder {
	q.Query += "GROUP BY "
	length := len(columns)
	for key, column := range columns {
		q.Query += fmt.Sprintf("%v", column)
		if key < length {
			q.Query += ", "
		}
	}

	return q
}

/*
	JOIN
*/

func (q *QueryBuilder) InnerJoin(table string) *QueryBuilder {
	q.Query += fmt.Sprintf("INNER JOIN %v ", table)
	return q
}

func (q *QueryBuilder) Join(table string) *QueryBuilder {
	return q.InnerJoin(table)
}

/*
	AS
*/

func (q *QueryBuilder) As(alias string) *QueryBuilder {
	q.Query += fmt.Sprintf("AS %v ", alias)
	return q
}

/*
	ON
*/

func (q *QueryBuilder) On(left string, right string) *QueryBuilder {
	q.Query += fmt.Sprintf("ON %v = %v ", left, right)
	return q
}

/**
UPDATE
*/

func (q *QueryBuilder) Update(table string) *QueryBuilder {
	q.Query += fmt.Sprintf("UPDATE %v ", table)
	return q
}

type SetOption struct {
	Column string
	Value  string
}

func (q *QueryBuilder) Set(values []SetOption) *QueryBuilder {
	q.Query += "SET "
	for i, value := range values {
		q.Query += fmt.Sprintf("%v = %v", value.Column, value.Value)
		if i < len(values)-1 {
			q.Query += ", "
		}
	}
	q.Query += " "
	return q
}

/**
INSERT
*/

func (q *QueryBuilder) InsertInto(table string, columns []string, values []string) *QueryBuilder {
	columnsJoined := strings.Join(columns, ", ")
	valuesJoined := strings.Join(values, ", ")
	q.Query += fmt.Sprintf("INSERT INTO %v (%v) VALUES (%v)", table, columnsJoined, valuesJoined)
	return q
}

/**
DELETE
*/

func (q *QueryBuilder) Delete(table string) *QueryBuilder {
	q.Query += fmt.Sprintf("DELETE FROM %v ", table)
	return q
}

package db

import (
	"fmt"
	"strings"
)

type QueryBuilder struct {
	Type  string
	Query string
}

func NewQueryBuilder(queryBuilderType string) *QueryBuilder {
	return &QueryBuilder{Type: queryBuilderType}
}

func (q *QueryBuilder) Get() string {
	return strings.TrimSpace(q.Query)
}

/* SELECT */
func (q *QueryBuilder) SelectOne(selectOptions string) *QueryBuilder {
	q.Query += fmt.Sprintf("SELECT %s", selectOptions)
	return q
}

func (q *QueryBuilder) SelectMany(selectOptions []string) *QueryBuilder {
	q.Query += "SELECT "
	for i, selectOption := range selectOptions {
		q.Query += selectOption

		if i < len(selectOptions)-1 {
			q.Query += ", "
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

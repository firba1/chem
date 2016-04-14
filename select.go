package chem

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

type SelectStmt struct {
	columns []Column
	filters []Filter
}

func Select(columnThings ...Columnser) SelectStmt {
	columns := make([]Column, 0, len(columnThings))
	for _, c := range columnThings {
		columns = append(columns, c.Columns()...)
	}

	return SelectStmt{columns: columns}
}

func (s SelectStmt) Where(filters ...Filter) SelectStmt {
	s.filters = filters
	return s
}

func toTableNames(columns []Column) []string {
	names := make(map[string]bool)
	for _, c := range columns {
		names[c.Table().Name()] = true
	}
	nameList := make([]string, 0, len(names))
	for name := range names {
		nameList = append(nameList, name)
	}
	return nameList
}

func toColumnExpressions(columns []Column) (out []string) {
	for _, c := range columns {
		out = append(out, c.toColumnExpression())
	}
	return
}

func toBooleanExpressions(filters []Filter) (out []string) {
	for _, filter := range filters {
		out = append(out, filter.toBooleanExpression())
	}
	return
}

func getAllBinds(filters []Filter) []interface{} {
	out := make([]interface{}, 0, len(filters))
	for _, f := range filters {
		out = append(out, f.binds()...)
	}
	return out
}

func flattenValues(values []interface{}) (out []interface{}) {
	for _, value := range values {
		reflection := reflect.ValueOf(value).Elem()
		reflectType := reflection.Type()
		switch reflectType.Kind() {
		case reflect.Struct:
			for i := 0; i < reflection.NumField(); i++ {
				out = append(out, reflection.Field(i).Addr().Interface())
			}
		default:
			out = append(out, value)
		}
	}
	return
}

func (s SelectStmt) One(tx *sql.Tx, values ...interface{}) error {
	stmt := fmt.Sprintf(
		"SELECT %v FROM %v WHERE %v",
		strings.Join(toColumnExpressions(s.columns), ", "),
		strings.Join(toTableNames(s.columns), ", "),
		strings.Join(toBooleanExpressions(s.filters), " AND "),
	)

	return tx.QueryRow(
		stmt,
		getAllBinds(s.filters)...,
	).Scan(flattenValues(values)...)
}
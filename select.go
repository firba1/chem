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

func (stmt SelectStmt) Where(filters ...Filter) SelectStmt {
	stmt.filters = append(stmt.filters, filters...)
	return stmt
}

func toTableNames(columns []Column) []string {
	names := make(map[string]bool)
	for _, column := range columns {
		names[column.Table().Name()] = true
	}
	nameList := make([]string, 0, len(names))
	for name := range names {
		nameList = append(nameList, name)
	}
	return nameList
}

func toColumnExpressions(columns []Column, withTableName bool) (out []string) {
	for _, column := range columns {
		out = append(out, column.toColumnExpression(withTableName))
	}
	return
}

func flattenValues(values []interface{}) []interface{} {
	out := make([]interface{}, 0, len(values))
	flattened := false

	for _, value := range values {
		reflection := reflect.ValueOf(value).Elem()
		reflectType := reflection.Type()
		switch reflectType.Kind() {
		case reflect.Struct:
			for i := 0; i < reflection.NumField(); i++ {
				// flatten this struct into pointers to each field of it
				out = append(out, reflection.Field(i).Addr().Interface())
			}
			flattened = true
		default:
			out = append(out, value)
		}
	}
	// if we ever flattened any structs, there could be nested structs, so let's recurse
	if flattened {
		return flattenValues(out)
	}
	// otherwise we're done
	return out
}

func makeWhereClause(f Filter, withTableNames bool) string {
	expression := f.toBooleanExpression(withTableNames)
	if expression == "" {
		return ""
	}
	return fmt.Sprintf("WHERE %v", expression)
}

func (stmt SelectStmt) prepareStmt(tx *sql.Tx) (*sql.Stmt, error) {
	tableNames := toTableNames(stmt.columns)
	fullyQualifyColumns := (len(tableNames) > 1)
	return tx.Prepare(
		strings.Join(
			filterStringSlice(
				fmt.Sprintf(
					"SELECT %v FROM %v",
					strings.Join(toColumnExpressions(stmt.columns, fullyQualifyColumns), ", "),
					strings.Join(tableNames, ", "),
				),
				makeWhereClause(AND(stmt.filters...), fullyQualifyColumns),
			),
			" ",
		),
	)
}

func (stmt SelectStmt) First(tx *sql.Tx, values ...interface{}) error {
	preparedStmt, err := stmt.prepareStmt(tx)
	if err != nil {
		return err
	}
	return preparedStmt.QueryRow(
		AND(stmt.filters...).binds()...,
	).Scan(flattenValues(values)...)
}

func (stmt SelectStmt) All(tx *sql.Tx, values ...interface{}) error {
	reflections := make([]reflect.Value, len(values))
	for i, value := range values {
		reflection := reflect.ValueOf(value).Elem()
		reflectType := reflection.Type()
		if reflectType.Kind() != reflect.Slice {
			return NonSliceError{
				Type: reflectType,
			}
		}
		reflections[i] = reflection
	}

	preparedStmt, err := stmt.prepareStmt(tx)
	if err != nil {
		return err
	}

	rows, err := preparedStmt.Query(
		AND(stmt.filters...).binds()...,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		// create new slice element value(s) to scan the database row into
		rowValues := make([]interface{}, len(reflections))
		for i, reflection := range reflections {
			rowValues[i] = reflect.New(reflection.Type().Elem()).Interface()
		}

		err = rows.Scan(flattenValues(rowValues)...)
		if err != nil {
			return err
		}

		for i, rowValue := range rowValues {
			reflections[i].Set(reflect.Append(reflections[i], reflect.ValueOf(rowValue).Elem()))
		}
	}

	return rows.Err()
}

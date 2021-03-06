package chem

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

type InsertStmt struct {
	table Table
}

func Insert(table Table) InsertStmt {
	return InsertStmt{table: table}
}

func binds(num int) []string {
	out := make([]string, num)
	for i := 0; i < num; i++ {
		out[i] = "?"
	}
	return out
}

func (stmt InsertStmt) Values(db DB, value interface{}) (sql.Result, error) {
	reflection := reflect.ValueOf(value)
	reflectedType := reflection.Type()

	columns := make([]string, reflectedType.NumField())
	values := make([]interface{}, reflectedType.NumField())
	for i := 0; i < reflectedType.NumField(); i++ {
		structField := reflectedType.Field(i)
		columns[i] = func(structField reflect.StructField) string {
			name := structField.Tag.Get("chem")
			if name == "" {
				name = structField.Name
			}
			return name
		}(structField)

		values[i] = reflection.Field(i).Interface()
	}

	queryString := fmt.Sprintf(
		"INSERT INTO %v (%v) VALUES (%v)",
		stmt.table.Name(),
		strings.Join(columns, ", "),
		strings.Join(binds(len(columns)), ", "),
	)
	preparedStmt, err := db.Prepare(queryString)
	if err != nil {
		return BadResult{err}, err
	}

	return preparedStmt.Exec(values...)
}

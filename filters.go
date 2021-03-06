package chem

import (
	"fmt"
	"strings"
)

type Filter interface {
	binds() []interface{}
	toBooleanExpression(withTableNames bool) string
}

const (
	equalsOperator             = "="
	notEqualsOperator          = "!="
	lessThanOperator           = "<"
	lessThanOrEqualOperator    = "<="
	greaterThanOperator        = ">"
	greaterThanOrEqualOperator = ">="
	likeOperator               = "LIKE"
)

type ValueFilter struct {
	column   Column
	operator string
	value    interface{}
}

func (f ValueFilter) toBooleanExpression(withTableNames bool) string {
	return fmt.Sprintf("%v %v ?", f.column.toColumnExpression(withTableNames), f.operator)
}

func (f ValueFilter) binds() []interface{} {
	return []interface{}{f.value}
}

type ColumnFilter struct {
	left     Column
	operator string
	right    Column
}

func (f ColumnFilter) toBooleanExpression(withTableNames bool) string {
	return fmt.Sprintf(
		"%v %v %v",
		f.left.toColumnExpression(withTableNames),
		f.operator,
		f.right.toColumnExpression(withTableNames),
	)
}

func (f ColumnFilter) binds() []interface{} {
	return []interface{}{}
}

type BooleanOperatorFilter struct {
	operator string
	filters  []Filter
}

func (f BooleanOperatorFilter) toBooleanExpression(withTableNames bool) string {
	expressionList := make([]string, len(f.filters))
	for i, filter := range f.filters {
		expressionList[i] = filter.toBooleanExpression(withTableNames)
	}
	expression := strings.Join(
		expressionList,
		fmt.Sprintf(" %v ", f.operator),
	)
	if expression == "" {
		return expression
	}
	// wrap the expression to make sure precedence is what user expects
	return fmt.Sprintf("( %v )", expression)
}

func (f BooleanOperatorFilter) binds() []interface{} {
	out := make([]interface{}, 0, len(f.filters))
	for _, filter := range f.filters {
		out = append(out, filter.binds()...)
	}
	return out
}

const (
	andOperator = "AND"
	orOperator  = "OR"
)

func AND(filters ...Filter) Filter {
	return BooleanOperatorFilter{
		operator: andOperator,
		filters:  filters,
	}
}

func OR(filters ...Filter) Filter {
	return BooleanOperatorFilter{
		operator: orOperator,
		filters:  filters,
	}
}

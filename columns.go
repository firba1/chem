package chem

import (
	"fmt"
	"sort"
)

type Column interface {
	Table() Table
	toColumnExpression(withTableName bool) string
}

type columns struct {
	columns *[]Column
	less    func(left, right Column) bool
}

func (cols columns) Len() int {
	return len(*cols.columns)
}

func (cols columns) Swap(i, j int) {
	(*cols.columns)[i], (*cols.columns)[j] = (*cols.columns)[j], (*cols.columns)[i]
}

func (cols columns) Less(i, j int) bool {
	return cols.less((*cols.columns)[i], (*cols.columns)[j])
}

func lessColumnsByUnqualifiedName(left, right Column) bool {
	return left.toColumnExpression(false) < right.toColumnExpression(false)
}

func sortColumns(cols []Column) []Column {
	copiedSlice := cols[:]
	sorter := columns{
		columns: &copiedSlice,
		less:    lessColumnsByUnqualifiedName,
	}
	sort.Sort(sorter)
	return *sorter.columns
}

type BaseColumn struct {
	Container Table
	Name      string
}

func (c BaseColumn) toColumnExpression(withTableName bool) string {
	if withTableName {
		return fmt.Sprintf("%v.%v", c.Container.Name(), c.Name)
	}
	return c.Name
}

func (c BaseColumn) Table() Table {
	return c.Container
}

func (c BaseColumn) Asc() Ordering {
	return ColumnOrdering{
		column:     c,
		descending: false,
	}
}

func (c BaseColumn) Desc() Ordering {
	return ColumnOrdering{
		column:     c,
		descending: true,
	}
}

type IntegerColumn struct {
	BaseColumn
}

func (c IntegerColumn) Equals(i int) Filter {
	return ValueFilter{
		column:   c,
		operator: equalsOperator,
		value:    i,
	}
}

type StringColumn struct {
	BaseColumn
}

func (c StringColumn) Equals(s string) Filter {
	return ValueFilter{
		column:   c,
		operator: equalsOperator,
		value:    s,
	}
}

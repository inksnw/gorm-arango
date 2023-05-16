package clause

import (
	"fmt"
	"strings"

	gormClause "gorm.io/gorm/clause"
)

type Sort struct {
	Columns    []gormClause.OrderByColumn
	Expression gormClause.Expression
}

func (sort Sort) Name() string {
	return "SORT"
}

func (sort Sort) Build(builder gormClause.Builder) {
	if sort.Expression != nil {
		sort.Expression.Build(builder)
		return
	}
	for idx, column := range sort.Columns {
		if idx > 0 {
			builder.WriteByte(',')
		}
		columnName := column.Column.Name

		if columnName == gormClause.PrimaryKey || columnName == gormClause.CurrentTable || columnName == gormClause.Associations {
			key := fmt.Sprintf("doc.%s", columnName)
			builder.WriteString(key)
			if column.Desc {
				builder.WriteString(" DESC")
			}
		} else {
			internalColumns := strings.Split(column.Column.Name, ",")
			for idy, internalCol := range internalColumns {
				if idy > 0 {
					builder.WriteString(", ")
				}
				key := fmt.Sprintf("doc.%s", strings.TrimSpace(internalCol))
				builder.WriteString(key)
				if column.Desc {
					builder.WriteString(" DESC")
				}
			}
		}
	}

}

func (sort Sort) MergeClause(clause *gormClause.Clause) {
	if v, ok := clause.Expression.(Sort); ok {
		for i := len(sort.Columns) - 1; i >= 0; i-- {
			if sort.Columns[i].Reorder {
				sort.Columns = sort.Columns[i:]
				clause.Expression = sort
				return
			}
		}

		copiedColumns := make([]gormClause.OrderByColumn, len(v.Columns))
		copy(copiedColumns, v.Columns)
		sort.Columns = append(copiedColumns, sort.Columns...)
	}

	clause.Expression = sort
}

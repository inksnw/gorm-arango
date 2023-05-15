package clause

import (
	"fmt"
	"log"
	"strings"
	"time"

	gormClause "gorm.io/gorm/clause"
)

type Filter struct {
	Exprs []gormClause.Expression
}

func (f Filter) Name() string {
	return "FILTER"
}

type FilterObj struct {
	Field    string
	Operator string
	Value    any
}

func parseFilter(expr gormClause.Expression, builder gormClause.Builder) (filter FilterObj) {

	switch e := expr.(type) {
	case gormClause.Eq:
		key := e.Column.(string)
		filter = FilterObj{Field: key, Operator: "==", Value: e.Value}

	case gormClause.Expr:
		conditions := strings.Split(strings.TrimSpace(e.SQL), "and")
		for idx, condition := range conditions {
			args := strings.Split(strings.TrimSpace(condition), " ")
			if len(args) != 3 {
				return
			}
			field, operator := args[0], args[1]
			if operator == "=" {
				operator = "=="
			}
			filter = FilterObj{Field: field, Operator: operator, Value: e.Vars[idx]}
		}
	default:
		expr.Build(builder)
	}

	return filter
}

func (f Filter) Build(builder gormClause.Builder) {
	var filterList []FilterObj
	for _, i := range f.Exprs {
		filter := parseFilter(i, builder)
		if filter.Field != "" {
			filterList = append(filterList, filter)
		}
	}
	var filterSlice []string
	for _, ins := range filterList {
		var sub string
		switch v := ins.Value.(type) {
		case string:
			sub = fmt.Sprintf("doc.%s %s '%s'", ins.Field, ins.Operator, v)
		case int:
			sub = fmt.Sprintf("doc.%s %s %d", ins.Field, ins.Operator, v)
		case bool:
			sub = fmt.Sprintf("doc.%s %s %t", ins.Field, ins.Operator, v)
		case time.Time:
			sub = fmt.Sprintf("DATE_TIMESTAMP(doc.%s) %s DATE_TIMESTAMP('%s')", ins.Field, ins.Operator, v.Format(time.RFC3339))
		default:
			log.Fatalf("not support operator %s for now ", ins.Operator)
		}

		filterSlice = append(filterSlice, sub)
	}
	sql := strings.Join(filterSlice, " AND ")
	builder.WriteString(sql)
}

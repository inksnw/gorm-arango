package clause

import (
	"fmt"
	"strconv"

	gormClause "gorm.io/gorm/clause"
)

type Limit struct {
	Limit  int
	Offset int
}

func (limit Limit) Name() string {
	return "LIMIT"
}

func (limit Limit) Build(builder gormClause.Builder) {
	//limit 2,5  offset2, limit5
	if limit.Offset > 0 || limit.Limit > 0 {
		builder.WriteString("LIMIT ")
	}

	if limit.Offset > 0 {
		var offset string
		if limit.Limit > 0 {
			offset = fmt.Sprintf("%s, ", strconv.Itoa(limit.Offset))
		} else {
			offset = fmt.Sprintf("%s ", strconv.Itoa(limit.Offset))
		}
		builder.WriteString(offset)
	}

	if limit.Limit > 0 {
		li := fmt.Sprintf("%s ", strconv.Itoa(limit.Limit))
		builder.WriteString(li)
	} else {
		builder.WriteString("250 ")
	}

}

func (limit Limit) MergeClause(clause *gormClause.Clause) {
	clause.Name = ""

	if v, ok := clause.Expression.(Limit); ok {
		if limit.Limit == 0 && v.Limit != 0 {
			limit.Limit = v.Limit
		}

		if limit.Offset == 0 && v.Offset > 0 {
			limit.Offset = v.Offset
		} else if limit.Offset < 0 {
			limit.Offset = 0
		}
	}

	clause.Expression = limit
}

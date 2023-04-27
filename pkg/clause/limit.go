package clause

import (
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

	if limit.Limit > 0 || limit.Offset > 0 {
		builder.WriteString("LIMIT ")
		if limit.Limit > 0 {
			builder.WriteString(strconv.Itoa(limit.Limit))
		} else {
			builder.WriteString("250")
		}
		builder.WriteString(strconv.Itoa(limit.Offset))
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

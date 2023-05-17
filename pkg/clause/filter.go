package clause

import (
	"fmt"
	"github.com/xwb1989/sqlparser"
	gormClause "gorm.io/gorm/clause"
	"strings"
	"time"
)

type Filter struct {
	Exprs []gormClause.Expression
}

func (f Filter) Name() string {
	return "FILTER"
}

func sqlToAqlCondition(expr sqlparser.Expr) (sql string) {

	switch sqlExpr := expr.(type) {
	case *sqlparser.ComparisonExpr:
		if sqlExpr.Operator == "=" {
			sqlExpr.Operator = "=="
		}
		column := sqlparser.String(sqlExpr.Left)
		value := sqlparser.String(sqlExpr.Right)
		if _, err := time.Parse(time.RFC3339, strings.Trim(value, "'")); err == nil {
			return fmt.Sprintf("DATE_TIMESTAMP(doc.%s) %s DATE_TIMESTAMP('%s')", column, sqlExpr.Operator, value)
		}

		return fmt.Sprintf("doc.%s %s %s", column, sqlExpr.Operator, value)
	case *sqlparser.AndExpr:
		return fmt.Sprintf("%s AND %s", sqlToAqlCondition(sqlExpr.Left), sqlToAqlCondition(sqlExpr.Right))
	case *sqlparser.OrExpr:
		return fmt.Sprintf("%s or %s", sqlToAqlCondition(sqlExpr.Left), sqlToAqlCondition(sqlExpr.Right))
	case *sqlparser.ParenExpr:
		return fmt.Sprintf("(%s)", sqlToAqlCondition(sqlExpr.Expr))
	default:
		return ""
	}
}

func parseFilter(expr gormClause.Expression, builder gormClause.Builder) (sql string) {

	switch e := expr.(type) {
	case gormClause.Eq:
		key := e.Column.(string)
		sql = fmt.Sprintf("doc.%s %s '%s'", key, "==", e.Value)
		return sql

	case gormClause.Expr:
		where := e.SQL
		for _, i := range e.Vars {
			switch i.(type) {
			case string:
				where = strings.Replace(where, "?", "'%v'", 1)
			default:
				where = strings.Replace(where, "?", "%v", 1)
			}
		}
		where = fmt.Sprintf(where, e.Vars...)
		fakeSql := fmt.Sprintf("select * from fake where %s", where)
		stmt, _ := sqlparser.Parse(fakeSql)
		selectStmt, _ := stmt.(*sqlparser.Select)
		sql = sqlToAqlCondition(selectStmt.Where.Expr)
		return sql

	default:
		expr.Build(builder)
	}

	return sql
}

func (f Filter) Build(builder gormClause.Builder) {
	var sqlList []string
	for _, i := range f.Exprs {
		sub := parseFilter(i, builder)
		if sub != "" {
			sqlList = append(sqlList, sub)
		}
	}
	sql := strings.Join(sqlList, " and ")
	builder.WriteString(sql)
}

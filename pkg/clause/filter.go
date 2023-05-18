package clause

import (
	"context"
	"fmt"
	"github.com/xwb1989/sqlparser"
	"gorm.io/gorm"
	gormClause "gorm.io/gorm/clause"
	"strings"
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
		values := sqlparser.String(sqlExpr.Right)
		if sqlExpr.Operator == "in" || sqlExpr.Operator == "not in" {
			values = strings.ReplaceAll(values, "(", "[")
			values = strings.ReplaceAll(values, ")", "]")
			sql = fmt.Sprintf("doc.%s %s %s", column, sqlExpr.Operator, values)
			return sql
		}

		if strings.HasPrefix(values, "'timestamp:") {
			values = strings.ReplaceAll(values, "'timestamp:", "")
			values = strings.ReplaceAll(values, "'", "")
			sql = fmt.Sprintf("DATE_TIMESTAMP(doc.%s) %s %s", column, sqlExpr.Operator, values)
			return sql
		}

		return fmt.Sprintf("doc.%s %s %s", column, sqlExpr.Operator, values)
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
	stmt1 := builder.(*gorm.Statement)
	stmt1.SQL.Reset()
	switch e := expr.(type) {
	case gormClause.Eq:
		key := e.Column.(string)
		sql = fmt.Sprintf("doc.%s %s '%s'", key, "==", e.Value)
		return sql

	case gormClause.Expr:
		e.Build(builder)
		where := stmt1.SQL.String()
		stmt1.SQL.Reset()
		fakeSql := fmt.Sprintf("select * from fake where %s", where)
		stmt, err := sqlparser.Parse(fakeSql)
		if err != nil {
			stmt1.Logger.Error(context.TODO(), err.Error(), where)
			return sql
		}
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

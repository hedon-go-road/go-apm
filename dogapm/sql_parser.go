package dogapm

import (
	"fmt"

	"github.com/xwb1989/sqlparser"
)

type sqlParser struct{}

var SQLParser = &sqlParser{}

func (p *sqlParser) parseTable(sql string) (tableName string, queryType int, err error, mutilTable bool) {
	queryType = sqlparser.Preview(sql)
	stmt, err := sqlparser.Parse(sql)
	if err != nil {
		return "", 0, fmt.Errorf("parse sql failed: %w", err), false
	}

	switch queryType {
	case sqlparser.StmtInsert:
		t := stmt.(*sqlparser.Insert).Table.Name
		return t.CompliantName(), sqlparser.StmtInsert, nil, false
	case sqlparser.StmtDelete:
		tExprs := stmt.(*sqlparser.Delete).TableExprs
		if len(tExprs) > 1 {
			return "", sqlparser.StmtDelete, nil, true
		}
		t := sqlparser.GetTableName(tExprs[0].(*sqlparser.AliasedTableExpr).Expr)
		return t.CompliantName(), sqlparser.StmtDelete, nil, false
	case sqlparser.StmtUpdate:
		tExprs := stmt.(*sqlparser.Update).TableExprs
		if len(tExprs) > 1 {
			return "", sqlparser.StmtUpdate, nil, true
		}
		t := sqlparser.GetTableName(tExprs[0].(*sqlparser.AliasedTableExpr).Expr)
		return t.CompliantName(), sqlparser.StmtUpdate, nil, false
	case sqlparser.StmtSelect:
		tExprs := stmt.(*sqlparser.Select).From
		if len(tExprs) > 1 {
			return "", sqlparser.StmtSelect, nil, true
		}
		t := sqlparser.GetTableName(tExprs[0].(*sqlparser.AliasedTableExpr).Expr)
		return t.CompliantName(), sqlparser.StmtSelect, nil, false
	}

	return "", 0, fmt.Errorf("unsupported query type: %d", queryType), false
}

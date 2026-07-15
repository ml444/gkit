package gorm

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const DefaultDeferredJoinOffsetThreshold = 1000

// DeferredJoinOffsetThreshold controls when deferred-join pagination activates.
var DeferredJoinOffsetThreshold = DefaultDeferredJoinOffsetThreshold

// CanDeferredJoin reports whether deferred-join pagination can be used.
func (gs *GormScope) CanDeferredJoin(offset int) (pk string, orderExpr string, ok bool) {
	return gs.canDeferredJoin(offset)
}

func (gs *GormScope) canDeferredJoin(offset int) (pk string, orderExpr string, ok bool) {
	if gs == nil || gs.gormDB == nil || gs.gormDB.Statement == nil {
		return "", "", false
	}
	if DeferredJoinOffsetThreshold <= 0 || offset < DeferredJoinOffsetThreshold {
		return "", "", false
	}

	stmt := gs.gormDB.Statement
	if stmt.Schema == nil {
		if gs.Model() != nil {
			if err := stmt.Parse(gs.Model()); err != nil {
				return "", "", false
			}
		} else if stmt.Model != nil {
			if err := stmt.Parse(stmt.Model); err != nil {
				return "", "", false
			}
		}
	}
	if stmt.Distinct {
		return "", "", false
	}
	if len(stmt.Joins) > 0 {
		return "", "", false
	}
	if _, hasGroup := stmt.Clauses["GROUP BY"]; hasGroup {
		return "", "", false
	}
	if _, hasHaving := stmt.Clauses["HAVING"]; hasHaving {
		return "", "", false
	}

	if stmt.Schema == nil || len(stmt.Schema.PrimaryFields) != 1 {
		return "", "", false
	}
	pk = stmt.Schema.PrimaryFields[0].DBName
	if pk == "" {
		return "", "", false
	}

	orderExpr, ok = deferredJoinOrderExpr(pk, stmt)
	if !ok {
		return "", "", false
	}
	return pk, orderExpr, true
}

func deferredJoinOrderExpr(pk string, stmt *gorm.Statement) (string, bool) {
	c, ok := stmt.Clauses["ORDER BY"]
	if !ok {
		return pk + " ASC", true
	}
	orderBy, ok := c.Expression.(clause.OrderBy)
	if !ok {
		return "", false
	}
	if len(orderBy.Columns) == 0 {
		return pk + " ASC", true
	}
	if len(orderBy.Columns) != 1 {
		return "", false
	}
	colName, desc, ok := parseOrderByColumn(orderBy.Columns[0])
	if !ok || !strings.EqualFold(colName, pk) {
		return "", false
	}
	if desc {
		return pk + " DESC", true
	}
	return pk + " ASC", true
}

func parseOrderByColumn(col clause.OrderByColumn) (name string, desc bool, ok bool) {
	if col.Column.Raw {
		parts := strings.Fields(col.Column.Name)
		if len(parts) == 0 {
			return "", false, false
		}
		name = strings.Trim(parts[0], "`\"")
		if len(parts) > 1 {
			desc = strings.EqualFold(parts[1], "DESC")
		}
		return name, desc, true
	}
	if col.Column.Name == "" {
		return "", false, false
	}
	return col.Column.Name, col.Desc, true
}

func (gs *GormScope) deferredJoinTable() string {
	if gs.gormDB.Statement.Schema != nil && gs.gormDB.Statement.Schema.Table != "" {
		return gs.gormDB.Statement.Schema.Table
	}
	return gs.gormDB.Statement.Table
}

func (gs *GormScope) findWithDeferredJoin(list any, limit, offset int, pk, orderExpr string) error {
	table := gs.deferredJoinTable()
	if table == "" {
		return gs.gormDB.Limit(limit).Offset(offset).Find(list).Error
	}

	quotedTable := gs.gormDB.Statement.Quote(table)
	quotedPK := gs.gormDB.Statement.Quote(pk)

	sub := gs.gormDB.Session(&gorm.Session{})
	clearPaginationClauses(sub.Statement)
	sub = sub.Select(pk).Order(orderExpr).Limit(limit).Offset(offset)

	joinSQL := fmt.Sprintf(
		"INNER JOIN (?) AS `_dbx_deferred` ON %s.%s = `_dbx_deferred`.%s",
		quotedTable, quotedPK, quotedPK,
	)

	outer := gs.gormDB.Session(&gorm.Session{})
	clearPaginationClauses(outer.Statement)
	if orderExpr != "" {
		outer = outer.Order(fmt.Sprintf("%s.%s %s", quotedTable, quotedPK, orderDirection(orderExpr)))
	}

	return outer.Joins(joinSQL, sub).Find(list).Error
}

func clearPaginationClauses(stmt *gorm.Statement) {
	delete(stmt.Clauses, "ORDER BY")
	delete(stmt.Clauses, "LIMIT")
	delete(stmt.Clauses, "OFFSET")
}

func orderDirection(orderExpr string) string {
	if strings.HasSuffix(strings.ToUpper(strings.TrimSpace(orderExpr)), " DESC") {
		return "DESC"
	}
	return "ASC"
}

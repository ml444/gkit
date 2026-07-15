package sqlx

import (
	"database/sql"
	"errors"
	"reflect"
	"testing"

	"github.com/ml444/gkit/dbx"
)

type compileRow struct {
	ID        int64 `json:"id"`
	DisplayID string
	private   string
}

func (compileRow) TableName() string { return "compile_rows" }

func TestCompileQueries(t *testing.T) {
	b := &dbx.QueryBuilder{
		Model: compileRow{}, Selects: []string{"id", "display_id"},
		Wheres:   []dbx.WhereClause{{Query: "id = ?", Args: []any{1}}},
		OrWheres: []dbx.WhereClause{{Query: "display_id = ?", Args: []any{"x"}}},
		OrderRaw: []string{"id DESC"}, Orders: []dbx.OrderColumn{{Field: "display_id"}},
		Limit: 2, Offset: 1, ForUpdate: true,
	}
	q, args, err := compileSelect(b)
	if err != nil || q != "SELECT id, display_id FROM compile_rows WHERE id = ? AND display_id = ? ORDER BY id DESC, display_id ASC LIMIT ? OFFSET ? FOR UPDATE" || !reflect.DeepEqual(args, []any{1, "x", 2, 1}) {
		t.Fatalf("compileSelect = %q %#v %v", q, args, err)
	}
	q, args, err = compileCount(b)
	if err != nil || q != "SELECT COUNT(*) FROM compile_rows WHERE id = ? AND display_id = ?" || !reflect.DeepEqual(args, []any{1, "x"}) {
		t.Fatalf("compileCount = %q %#v %v", q, args, err)
	}
	if q, args := compileWhere(&dbx.QueryBuilder{}); q != "" || args != nil {
		t.Fatalf("compileWhere empty = %q %#v", q, args)
	}
	if q := compileOrder(&dbx.QueryBuilder{}); q != "" {
		t.Fatalf("compileOrder empty = %q", q)
	}
	q, args, err = compileInsert(&dbx.QueryBuilder{Table: "compile_rows"}, compileRow{ID: 1, DisplayID: "x"})
	if err != nil || q != "INSERT INTO compile_rows (id,display_i_d) VALUES (?,?)" || !reflect.DeepEqual(args, []any{int64(1), "x"}) {
		t.Fatalf("compileInsert = %q %#v %v", q, args, err)
	}
	q, args, err = compileUpdate(&dbx.QueryBuilder{Table: "compile_rows", Wheres: []dbx.WhereClause{{Query: "id = ?", Args: []any{1}}}}, map[string]any{"name": "x"})
	if err != nil || q == "" || len(args) != 2 {
		t.Fatalf("compileUpdate = %q %#v %v", q, args, err)
	}
	q, args, err = compileUpdateColumn(&dbx.QueryBuilder{Table: "compile_rows", Wheres: []dbx.WhereClause{{Query: "id = ?", Args: []any{1}}}}, "name", "x")
	if err != nil || q != "UPDATE compile_rows SET name = ? WHERE id = ?" || !reflect.DeepEqual(args, []any{"x", 1}) {
		t.Fatalf("compileUpdateColumn = %q %#v %v", q, args, err)
	}
	q, args, err = compileUpdateColumn(&dbx.QueryBuilder{Table: "compile_rows", IncrColumn: "score", IncrValue: -2}, "ignored", nil)
	if err != nil || q != "UPDATE compile_rows SET score = COALESCE(score, 0) - ?" || !reflect.DeepEqual(args, []any{int64(2)}) {
		t.Fatalf("compileUpdateColumn increment = %q %#v %v", q, args, err)
	}
	q, args, err = compileDelete(&dbx.QueryBuilder{Table: "compile_rows"})
	if err != nil || q != "DELETE FROM compile_rows" || args != nil {
		t.Fatalf("compileDelete = %q %#v %v", q, args, err)
	}
}

func TestCompilerHelpers(t *testing.T) {
	if tableName(&dbx.QueryBuilder{Table: "explicit"}) != "explicit" || tableName(&dbx.QueryBuilder{Model: compileRow{}}) != "compile_rows" {
		t.Fatal("tableName did not prefer table/model")
	}
	if tableNameFromModel(&compileRow{}) != "compile_row" || tableNameFromModel(1) != "" || tableNameFromModel(nil) != "" {
		t.Fatal("tableNameFromModel mismatch")
	}
	cols, vals, err := structColumns(&compileRow{ID: 1, DisplayID: "x"})
	if err != nil || !reflect.DeepEqual(cols, []string{"id", "display_i_d"}) || !reflect.DeepEqual(vals, []any{int64(1), "x"}) {
		t.Fatalf("structColumns = %#v %#v %v", cols, vals, err)
	}
	cols, vals, err = structColumns(map[string]any{"x": 1})
	if err != nil || len(cols) != 1 || cols[0] != "x" || vals[0] != 1 {
		t.Fatalf("mapColumns = %#v %#v %v", cols, vals, err)
	}
	if _, _, err := structColumns(1); err == nil {
		t.Fatal("structColumns accepted scalar")
	}
	field, _ := reflect.TypeOf(compileRow{}).FieldByName("DisplayID")
	if columnName(field) != "display_i_d" || camelToSnake("HTTPServer") != "h_t_t_p_server" {
		t.Fatal("name conversion mismatch")
	}
	want := errors.New("x")
	if !errors.Is(mapErr(want), want) || !errors.Is(mapErr(sql.ErrNoRows), dbx.ErrRecordNotFound) || mapErr(nil) != nil {
		t.Fatal("mapErr mismatch")
	}
}

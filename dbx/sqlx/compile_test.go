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
	if err != nil || q != "SELECT id, display_id FROM compile_rows WHERE id = ? AND (display_id = ?) ORDER BY id DESC, display_id ASC LIMIT ? OFFSET ? FOR UPDATE" || !reflect.DeepEqual(args, []any{1, "x", 2, 1}) {
		t.Fatalf("compileSelect = %q %#v %v", q, args, err)
	}
	q, args, err = compileCount(b)
	if err != nil || q != "SELECT COUNT(*) FROM compile_rows WHERE id = ? AND (display_id = ?)" || !reflect.DeepEqual(args, []any{1, "x"}) {
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

type UpsertBase struct {
	TenantID int64 `db:"tenant_id,pk"`
	Ignored  int64 `db:"-"`
}

type upsertRow struct {
	*UpsertBase
	ID      int64
	Name    string `db:"name"`
	private string
}

type upsertValueEmbeddedRow struct {
	UpsertBase
	Name string `db:"name"`
}

type onlyPKRow struct {
	ID int64 `db:"id,pk"`
}

type noPKRow struct {
	Name string `db:"name"`
}

func TestCompileUpsertBranches(t *testing.T) {
	b := &dbx.QueryBuilder{Table: "upsert_rows"}
	q, args, err := compileUpsert(b, upsertRow{UpsertBase: &UpsertBase{TenantID: 7, Ignored: 9}, ID: 1, Name: "a"}, "mysql")
	if err != nil {
		t.Fatal(err)
	}
	if q != "INSERT INTO upsert_rows (tenant_id, id, name) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE name=VALUES(name)" {
		t.Fatalf("mysql upsert = %q", q)
	}
	if !reflect.DeepEqual(args, []any{int64(7), int64(1), "a"}) {
		t.Fatalf("mysql args = %#v", args)
	}
	q, args, err = compileUpsert(b, upsertRow{UpsertBase: &UpsertBase{TenantID: 7}, ID: 1, Name: "a"}, "sqlite")
	if err != nil {
		t.Fatal(err)
	}
	if q != "INSERT INTO upsert_rows (tenant_id, id, name) VALUES (?, ?, ?) ON CONFLICT (tenant_id, id) DO UPDATE SET name=EXCLUDED.name" || len(args) != 3 {
		t.Fatalf("sqlite upsert = %q %#v", q, args)
	}
	q, args, err = compileUpsert(b, upsertValueEmbeddedRow{UpsertBase: UpsertBase{TenantID: 7}, Name: "a"}, "postgres")
	if err != nil {
		t.Fatal(err)
	}
	if q != "INSERT INTO upsert_rows (tenant_id, name) VALUES (?, ?) ON CONFLICT (tenant_id) DO UPDATE SET name=EXCLUDED.name" || !reflect.DeepEqual(args, []any{int64(7), "a"}) {
		t.Fatalf("embedded value upsert = %q %#v", q, args)
	}
	q, _, err = compileUpsert(b, onlyPKRow{ID: 1}, "postgres")
	if err != nil || q != "INSERT INTO upsert_rows (id) VALUES (?) ON CONFLICT DO NOTHING" {
		t.Fatalf("postgres do nothing = %q %v", q, err)
	}
	q, _, err = compileUpsert(b, onlyPKRow{ID: 1}, "mysql")
	if err != nil || q != "INSERT INTO upsert_rows (id) VALUES (?) ON DUPLICATE KEY UPDATE id=id" {
		t.Fatalf("mysql do nothing = %q %v", q, err)
	}
	if _, _, err := compileUpsert(b, onlyPKRow{ID: 1}, "oracle"); err == nil {
		t.Fatal("unsupported do-nothing dialect should fail")
	}
	if _, _, err := compileUpsert(b, noPKRow{Name: "a"}, "sqlite"); err == nil {
		t.Fatal("sqlite upsert without primary key should fail")
	}
	if _, _, err := compileUpsert(b, noPKRow{Name: "a"}, "oracle"); err == nil {
		t.Fatal("unsupported upsert dialect should fail")
	}
	if _, _, _, err := parseModelForUpsert((*upsertRow)(nil)); err == nil {
		t.Fatal("nil pointer should fail")
	}
	if _, _, _, err := parseModelForUpsert(1); err == nil {
		t.Fatal("scalar should fail")
	}
	type emptyStruct struct {
		private int
	}
	if _, _, _, err := parseModelForUpsert(emptyStruct{}); err == nil {
		t.Fatal("empty struct should fail")
	}
	var cols []string
	var vals []any
	var pks []string
	if err := extractStructFields(reflect.ValueOf((*UpsertBase)(nil)), &cols, &vals, &pks); err != nil || len(cols) != 0 {
		t.Fatalf("nil embedded pointer = %#v %v", cols, err)
	}
	if err := extractStructFields(reflect.ValueOf(1), &cols, &vals, &pks); err != nil {
		t.Fatalf("scalar extract should be ignored: %v", err)
	}
}

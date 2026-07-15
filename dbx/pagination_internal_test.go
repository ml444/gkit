package dbx

import (
	"reflect"
	"testing"
)

func TestPaginationCursorAndColumnHelpers(t *testing.T) {
	type tagged struct {
		GormName string `gorm:"column:external_name"`
		JSONName string `json:"json_name,omitempty"`
		Plain    int
	}
	v := reflect.ValueOf(tagged{GormName: "gorm", JSONName: "json", Plain: 7})
	if got := fieldByColumn(v, "external_name"); !got.IsValid() || got.String() != "gorm" {
		t.Fatalf("gorm column = %v", got)
	}
	if got := fieldByColumn(v, "plain"); !got.IsValid() || got.Int() != 7 {
		t.Fatalf("snake column = %v", got)
	}
	if got := fieldByColumn(v, "Plain"); !got.IsValid() || got.Int() != 7 {
		t.Fatalf("pascal column = %v", got)
	}
	if got := fieldByColumn(v, "missing"); got.IsValid() {
		t.Fatalf("missing column = %v", got)
	}
	if columnNameFromTag(reflect.TypeOf(tagged{}).Field(0)) != "external_name" {
		t.Fatal("gorm column tag missing")
	}
	if columnNameFromTag(reflect.TypeOf(struct{ Value string }{}).Field(0)) != "" {
		t.Fatal("untagged field has a column")
	}
	if snakeToPascal("a__B_c") != "ABC" {
		t.Fatalf("snakeToPascal = %q", snakeToPascal("a__B_c"))
	}
	for _, tc := range []struct {
		value any
		want  string
	}{
		{uint(1), "1"},
		{int64(-2), "-2"},
		{"text", "text"},
		{true, ""},
	} {
		got := formatScrollCursor(reflect.ValueOf(tc.value))
		if got != tc.want {
			t.Fatalf("formatScrollCursor(%T) = %q", tc.value, got)
		}
	}
}

type fieldByColumnModel struct {
	ID        uint64 `gorm:"primaryKey"`
	FooSar    uint64
	FoosAr    uint64
	CreatedAt uint32
	CustomCol uint64 `gorm:"column:custom_foo_sar"`
}

type fieldByColumnIDModel struct {
	ID uint64
}

func TestFieldByColumn(t *testing.T) {
	m := fieldByColumnModel{
		ID:        10,
		FooSar:    1,
		FoosAr:    2,
		CreatedAt: 3,
		CustomCol: 4,
	}
	v := reflect.ValueOf(m)

	tests := []struct {
		column string
		want   uint64
	}{
		{column: "id", want: 10},
		{column: "foo_sar", want: 1},
		{column: "foos_ar", want: 2},
		{column: "created_at", want: 3},
		{column: "custom_foo_sar", want: 4},
	}
	for _, tt := range tests {
		t.Run(tt.column, func(t *testing.T) {
			got := fieldByColumn(v, tt.column)
			if !got.IsValid() {
				t.Fatalf("fieldByColumn(%q) returned invalid value", tt.column)
			}
			if got.Uint() != tt.want {
				t.Fatalf("fieldByColumn(%q) = %d, want %d", tt.column, got.Uint(), tt.want)
			}
		})
	}

	idV := reflect.ValueOf(fieldByColumnIDModel{ID: 99})
	got := fieldByColumn(idV, "id")
	if !got.IsValid() || got.Uint() != 99 {
		t.Fatalf("fieldByColumn(id) on ID field = %v, want 99", got)
	}
}

func TestSnakeToPascal(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{in: "foo_sar", want: "FooSar"},
		{in: "foos_ar", want: "FoosAr"},
		{in: "created_at", want: "CreatedAt"},
		{in: "id", want: "Id"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := snakeToPascal(tt.in); got != tt.want {
				t.Fatalf("snakeToPascal(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestScope_HandlePagination(t *testing.T) {
	scope := &Scope{}
	t.Run("zero pagination", func(t *testing.T) {
		opt := scope.HandlePagination(0, 0)
		if opt.Size != uint32(DefaultLimit) {
			t.Errorf("%d != %d", opt.Size, DefaultLimit)
		}
	})

	t.Run("size greater than max limit", func(t *testing.T) {
		opt := scope.HandlePagination(1, uint32(MaxLimit)+1)
		if opt.Size != uint32(MaxLimit) {
			t.Errorf("%d != %d", opt.Size, MaxLimit)
		}
	})
}

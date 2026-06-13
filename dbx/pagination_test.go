package dbx

import (
	"reflect"
	"testing"

	"gorm.io/gorm"

	"github.com/ml444/gkit/dbx/pagination"
)

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

func TestScope_PaginationQuery(t *testing.T) {
	type fields struct {
		DB *gorm.DB
	}
	type args struct {
		opt  *pagination.Pagination
		list *[]*testOrmModel
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *pagination.Pagination
		wantErr bool
	}{
		{
			name: "normal case",
			fields: fields{
				DB: testGetDB(),
			},
			args: args{
				opt:  &pagination.Pagination{Page: 1, Size: 10},
				list: &[]*testOrmModel{},
			},
			want:    &pagination.Pagination{Page: 1, Size: 10},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewScope(testGetDB(), &testOrmModel{})
			got, err := s.PaginationQuery(tt.args.list, tt.args.opt.Page, tt.args.opt.Size)
			if (err != nil) != tt.wantErr {
				t.Errorf("Scope.PaginationQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.Page != tt.want.Page || got.Size != tt.want.Size {
				t.Errorf("Scope.PaginationQuery() = %v, want %v", got, tt.want)
			}
			if !got.SkipCount && got.Total > 0 {
				if got.Total > int64(got.Size) && len(*tt.args.list) != int(got.Size) {
					t.Errorf("find err: list len: %d, want: %d", len(*tt.args.list), got.Size)
				}
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

	t.Run("size less than 0", func(t *testing.T) {
		opt := scope.HandlePagination(1, 0)
		if opt.Size != uint32(DefaultLimit) {
			t.Errorf("%d != %d", opt.Size, DefaultLimit)
		}
	})

	t.Run("size greater than max limit", func(t *testing.T) {
		opt := &pagination.Pagination{
			Size: uint32(MaxLimit) + 1,
		}
		opt = scope.HandlePagination(opt.Page, opt.Size)
		if opt.Size != uint32(MaxLimit) {
			t.Errorf("%d != %d", opt.Size, MaxLimit)
		}
	})

	t.Run("normal case", func(t *testing.T) {
		opt := &pagination.Pagination{
			Size: 100,
		}
		opt = scope.HandlePagination(opt.Page, opt.Size)
		if opt.Size != 100 {
			t.Errorf("%d != %d", opt.Size, 100)
		}
	})
}

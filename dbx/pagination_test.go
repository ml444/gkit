package dbx

import (
	"reflect"
	"testing"

	"gorm.io/gorm"

	"github.com/ml444/gkit/dbx/pagination"
)

func TestScope_PaginationQuery(t *testing.T) {
	type fields struct {
		DB *gorm.DB
	}
	type args struct {
		opt  *pagination.Pagination
		list interface{}
	}
	opt := &pagination.Pagination{
		Page: 1,
		Size: 10,
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
				DB: &gorm.DB{},
			},
			args: args{
				opt:  opt,
				list: nil,
			},
			want:    &pagination.Pagination{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Scope{
				DB: tt.fields.DB,
			}
			got, err := s.PaginationQuery(tt.args.opt, tt.args.list)
			if (err != nil) != tt.wantErr {
				t.Errorf("Scope.PaginationQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Scope.PaginationQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScope_HandlePagination(t *testing.T) {
	scope := &Scope{}
	t.Run("nil pagination", func(t *testing.T) {
		opt := scope.HandlePagination(nil)
		if opt.Size != DefaultLimit {
			t.Errorf("%d != %d", opt.Size, DefaultLimit)
		}
	})

	t.Run("size less than 0", func(t *testing.T) {
		opt := &pagination.Pagination{
			Size: 0,
		}
		opt = scope.HandlePagination(opt)
		if opt.Size != DefaultLimit {
			t.Errorf("%d != %d", opt.Size, DefaultLimit)
		}
	})

	t.Run("size greater than max limit", func(t *testing.T) {
		opt := &pagination.Pagination{
			Size: MaxLimit + 1,
		}
		opt = scope.HandlePagination(opt)
		if opt.Size != MaxLimit {
			t.Errorf("%d != %d", opt.Size, MaxLimit)
		}
	})

	t.Run("normal case", func(t *testing.T) {
		opt := &pagination.Pagination{
			Size: 100,
		}
		opt = scope.HandlePagination(opt)
		if opt.Size != 100 {
			t.Errorf("%d != %d", opt.Size, 100)
		}
	})
}

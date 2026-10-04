package dbx

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

type countErrDriver struct{ stubDriver }

func (countErrDriver) Count(ctx context.Context, b *QueryBuilder) (int64, error) {
	return 0, errors.New("count boom")
}

type firstErrDriver struct{ stubDriver }

func (firstErrDriver) First(ctx context.Context, b *QueryBuilder, dest any) error {
	return errors.New("first boom")
}

func TestPaginationCountErrorAndExistError(t *testing.T) {
	s := NewScope(stubTxConn{d: countErrDriver{}}, &testRow{})
	var rows []testRow
	if _, err := s.PaginationQueryWithOpt(&rows, nil); err == nil {
		t.Fatal("count error")
	}
	s = NewScope(stubTxConn{d: firstErrDriver{}}, &testRow{})
	if _, err := s.Exist(); err == nil {
		t.Fatalf("exist count error")
	}
}

func TestWhereNonMapFallback(t *testing.T) {
	s := NewScope(stubTxConn{d: stubDriver{}}, &testRow{}).Where(42, "x")
	if len(s.Builder().Wheres) == 0 {
		t.Fatal("where missing")
	}
}

func TestResetTimeAndProtoUpsert(t *testing.T) {
	msg := mustDynamicUpsertMessage(t)
	stub := &protoUpsertStub{msg: msg}
	s := NewScope(stubTxConn{d: stubDriver{}}, &testRow{})
	s.resetTime = true
	s.ResetSysDateTimeField(stub)
	if err := s.Create(stub); err != nil {
		t.Fatal(err)
	}
	if err := s.Save(stub); err != nil {
		t.Fatal(err)
	}
	if err := s.Update(map[string]any{"name": "x"}); err != nil {
		t.Fatal(err)
	}
}

type protoUpsertStub struct {
	msg protoreflect.Message
}

func (p *protoUpsertStub) ProtoReflect() protoreflect.Message { return p.msg }
func (p *protoUpsertStub) GetCreatedAt() int64                { return 1 }
func (p *protoUpsertStub) GetUpdatedAt() int64                { return 1 }
func (p *protoUpsertStub) GetDeletedAt() int64                { return 0 }

func mustDynamicUpsertMessage(t *testing.T) protoreflect.Message {
	t.Helper()
	file := &descriptorpb.FileDescriptorProto{
		Name:    proto.String("dbx_upsert.proto"),
		Package: proto.String("dbx.test"),
		Syntax:  proto.String("proto3"),
		MessageType: []*descriptorpb.DescriptorProto{{
			Name: proto.String("Upsert"),
			Field: []*descriptorpb.FieldDescriptorProto{
				{
					Name:     proto.String("created_at"),
					Number:   proto.Int32(1),
					Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
					Type:     descriptorpb.FieldDescriptorProto_TYPE_UINT32.Enum(),
					JsonName: proto.String("created_at"),
				},
				{
					Name:     proto.String("updated_at"),
					Number:   proto.Int32(2),
					Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
					Type:     descriptorpb.FieldDescriptorProto_TYPE_UINT32.Enum(),
					JsonName: proto.String("updated_at"),
				},
				{
					Name:     proto.String("deleted_at"),
					Number:   proto.Int32(3),
					Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
					Type:     descriptorpb.FieldDescriptorProto_TYPE_UINT32.Enum(),
					JsonName: proto.String("deleted_at"),
				},
			},
		}},
	}
	fd, err := protodesc.NewFile(file, nil)
	if err != nil {
		t.Fatal(err)
	}
	md := fd.Messages().Get(0)
	msg := dynamicpb.NewMessage(md)
	fields := md.Fields()
	msg.Set(fields.ByName("created_at"), protoreflect.ValueOfUint32(9))
	msg.Set(fields.ByName("updated_at"), protoreflect.ValueOfUint32(9))
	msg.Set(fields.ByName("deleted_at"), protoreflect.ValueOfUint32(9))
	return msg
}

func TestScrollPointerLastAndPascalColumn(t *testing.T) {
	d := &seedFindDriver{rows: []map[string]any{{"id": int64(1), "name": "a", "age": int64(1), "deleted_at": uint32(0)}}}
	s := NewScope(stubTxConn{d: d}, &testRow{})
	var rows []*testRow
	if _, err := s.ScrollQuery(&rows, "id", "0", 10, false); err != nil {
		t.Fatal(err)
	}
	type weird struct {
		FooBar int `json:""`
	}
	v := reflect.ValueOf(weird{FooBar: 7})
	// double underscore skips camelToSnake match then hits snakeToPascal exact name
	if got := fieldByColumn(v, "foo__bar"); !got.IsValid() || got.Int() != 7 {
		t.Fatalf("pascal path = %v", got)
	}
}

type seedFindDriver struct {
	stubDriver
	rows []map[string]any
}

func (d seedFindDriver) Find(ctx context.Context, b *QueryBuilder, dest any) error {
	return assignTestRows(dest, d.rows)
}

func assignTestRows(dest any, rows []map[string]any) error {
	dv := reflect.ValueOf(dest)
	if dv.Kind() != reflect.Ptr {
		return fmt.Errorf("dest ptr")
	}
	sv := dv.Elem()
	elemType := sv.Type().Elem()
	isPtr := elemType.Kind() == reflect.Ptr
	if isPtr {
		elemType = elemType.Elem()
	}
	out := reflect.MakeSlice(sv.Type(), 0, len(rows))
	for _, row := range rows {
		val := reflect.New(elemType)
		rv := val.Elem()
		for i := 0; i < rv.NumField(); i++ {
			f := rv.Type().Field(i)
			col := f.Name
			if j := f.Tag.Get("json"); j != "" {
				col = j
			}
			if v, ok := row[col]; ok {
				rv.Field(i).Set(reflect.ValueOf(v).Convert(rv.Field(i).Type()))
			}
		}
		if isPtr {
			out = reflect.Append(out, val)
		} else {
			out = reflect.Append(out, val.Elem())
		}
	}
	sv.Set(out)
	return nil
}

type softDeleteCoverageRow struct {
	ID        int64 `json:"id"`
	DeletedAt int64 `json:"deleted_at"`
}

func (s softDeleteCoverageRow) GetDeletedAt() int64 { return s.DeletedAt }

func TestScopeRemainingCoverageBranches(t *testing.T) {
	_, conn := testConn()
	s := NewScope(conn, softDeleteCoverageRow{})
	if err := s.Delete("id = ?", int64(1)); err != nil {
		t.Fatalf("soft delete with conds: %v", err)
	}
	q := NewScope(conn, &testRow{}).Query(&QueryOpts{Like: map[string]string{"name": "a"}, IsLikePrefix: false})
	if len(q.Builder().Wheres) == 0 {
		t.Fatal("Query Like should add a where clause")
	}
}

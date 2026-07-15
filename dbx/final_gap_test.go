package dbx

import (
	"context"
	"testing"
)

func TestNeedEncryptWarnAndBatchInterface(t *testing.T) {
	d := stubDriver{}
	conn := stubTxConn{d: d}
	_ = NewT[testRow](func() Conn { return conn }, func(t *T) {
		t.NeedEncrypt = true
	})

	repo := NewT[testRow](func() Conn { return conn })
	var wrapped any = []*testRow{{ID: 1, Name: "a"}}
	if err := repo.BatchCreate(context.Background(), wrapped); err != nil {
		t.Fatal(err)
	}
}

type softDel2 struct {
	ID        int64 `json:"id"`
	DeletedAt int64 `json:"deleted_at"`
}

func (softDel2) TableName() string     { return "sr2" }
func (s softDel2) GetDeletedAt() int64 { return s.DeletedAt }

func TestSoftDeleteWithConds2(t *testing.T) {
	s := NewScope(stubTxConn{d: stubDriver{}}, softDel2{})
	if err := s.Delete("id = ?", int64(1)); err != nil {
		t.Fatal(err)
	}
}

type zeroUpd struct{ stubDriver }

func (z zeroUpd) Update(ctx context.Context, b *QueryBuilder, v any) (int64, error) {
	return 0, nil
}

func TestUpdateZeroRowsWarn(t *testing.T) {
	s := NewScope(stubTxConn{d: zeroUpd{}}, &testRow{})
	if err := s.Update(map[string]any{"name": "x"}); err != nil {
		t.Fatal(err)
	}
}

func TestQueryLikeNonPrefix(t *testing.T) {
	s := NewScope(stubTxConn{d: stubDriver{}}, &testRow{}).Query(&QueryOpts{
		Like:         map[string]string{"name": "a"},
		IsLikePrefix: false,
	})
	if len(s.Builder().Wheres) == 0 {
		t.Fatal("like where missing")
	}
}

func TestWhereGenericMap(t *testing.T) {
	s := NewScope(stubTxConn{d: stubDriver{}}, &testRow{}).Where(map[any]any{"id": int64(1)})
	if len(s.Builder().Wheres) == 0 {
		t.Fatal("generic map where")
	}
}

func TestTxConnDriverWithoutContext(t *testing.T) {
	conn := stubTxConn{d: noContextDriver{}}
	err := ScopeTxGo(context.Background(), conn, func() (any, func(*Scope) error) {
		return &testRow{}, func(s *Scope) error {
			if s.Conn().Driver(context.Background()) == nil {
				t.Fatal("nil driver")
			}
			return nil
		}
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestParseEmptyScrollCursor(t *testing.T) {
	v, err := parseScrollCursor("")
	if err != nil || v != nil {
		t.Fatalf("empty cursor = %v, %v", v, err)
	}
}

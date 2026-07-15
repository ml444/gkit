package dbx

import (
	"database/sql"
	"errors"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ml444/gkit/errorx"
)

func TestMapDriverError(t *testing.T) {
	if MapDriverError(nil) != nil {
		t.Fatal("nil error should remain nil")
	}
	if !errors.Is(MapDriverError(sql.ErrNoRows), ErrRecordNotFound) {
		t.Fatal("sql.ErrNoRows should map to ErrRecordNotFound")
	}
	other := errors.New("other")
	if MapDriverError(other) != other {
		t.Fatal("other error should be unchanged")
	}
}

func TestNotFoundErrors(t *testing.T) {
	got := GetNotFoundErr(errors.New("gone"))
	if got.GetCode() != errorx.ErrCodeRecordNotFoundSys || got.GetMessage() != "gone" {
		t.Fatalf("unexpected error: %#v", got)
	}
	if !IsNotFoundErr(errorx.New(88), 88) || !IsNotFoundErr(GetNotFoundErr(errors.New("x")), 0) {
		t.Fatal("errorx not found errors should match")
	}
	if !IsNotFoundErr(status.Error(codes.NotFound, "gone"), 0) {
		t.Fatal("grpc NotFound should match")
	}
	if IsNotFoundErr(errors.New("no"), 88) {
		t.Fatal("ordinary error should not match")
	}
	if !IsUpdateRowAffectedZero(ErrUpdateRowAffectedZero) || IsUpdateRowAffectedZero(errors.New("no")) {
		t.Fatal("update-zero classification incorrect")
	}
}

func TestIsDuplicateErr(t *testing.T) {
	for _, err := range []error{errors.New("Duplicate entry"), errors.New("duplicate key value violates unique constraint"), errors.New("UNIQUE constraint failed: test_rows.name")} {
		if !IsDuplicateErr(err) {
			t.Fatalf("expected duplicate for %v", err)
		}
	}
	if IsDuplicateErr(nil) || IsDuplicateErr(errors.New("other error")) {
		t.Fatal("unexpected duplicate match")
	}
}

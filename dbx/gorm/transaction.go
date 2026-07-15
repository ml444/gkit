package gorm

import (
	"context"
	"errors"

	"github.com/ml444/gkit/dbx"
	"gorm.io/gorm"
)

// TxGo runs fn inside a GORM transaction (compat helper for *gorm.DB callbacks).
func TxGo(ctx context.Context, conn dbx.Conn, fn func(tx *gorm.DB) error) error {
	return dbx.TxGo(ctx, conn, func(d dbx.Driver) error {
		tx, ok := RawDB(d)
		if !ok {
			return errors.New("expected gorm driver")
		}
		return fn(tx)
	})
}

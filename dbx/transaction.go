package dbx

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/ml444/gkit/log"
	"gorm.io/gorm"
)

type (
	TxHandler  func(tx *gorm.DB) error
	TxCallback func() (model interface{}, execute func(scope *Scope) error)
)

func TxGo(ctx context.Context, db *gorm.DB, executes ...TxHandler) (err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Transaction PANIC: %v", r)
			log.Errorf("Stack trace:\n%s", string(debug.Stack()))
			err = fmt.Errorf("Transaction PANIC: %v", r)
		}
	}()
	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, execute := range executes {
			if err := execute(tx); err != nil {
				return err
			}
		}
		return nil
	})
	return
}

func ScopeTxGo(ctx context.Context, db *gorm.DB, callbacks ...TxCallback) (err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Transaction PANIC: %v", r)
			log.Errorf("Stack trace:\n%s", string(debug.Stack()))
			err = fmt.Errorf("Transaction PANIC: %v", r)
		}
	}()

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, callback := range callbacks {
			model, exec := callback()
			scope := NewScope(tx, model)
			if err := exec(scope); err != nil {
				return err
			}
		}
		return nil
	})
}

func TxCreateMultiModels(ctx context.Context, db *gorm.DB, models ...any) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, m := range models {
			err := tx.Create(m).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
}


type TxItem interface {
	Preload(tx *gorm.DB) error
	Execute(tx *gorm.DB) error
}

func TxItems(ctx context.Context, db *gorm.DB, items ...TxItem) (err error) {
	if len(items) == 0 {
		return nil
	}
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Transaction PANIC: %v", r)
			log.Errorf("Stack trace:\n%s", string(debug.Stack()))
			err = fmt.Errorf("Transaction PANIC: %v", r)
		}
	}()

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) (err error) {
		for _, item := range items {
			if err := item.Preload(tx); err != nil {
				return err
			}
		}
		for _, item := range items {
			if err := item.Execute(tx); err != nil {
				return err
			}
		}
		return nil
	})
}

func NewInsertItem(models any) *InsertItem {
	return &InsertItem{
		Models: models,
	}
}

type InsertItem struct {
	Models any
}
func (i *InsertItem) Preload(tx *gorm.DB) error {
	return nil
}
func (i *InsertItem) Execute(tx *gorm.DB) error {
	if i.Models == nil {
		return nil
	}
	return tx.Create(i.Models).Error
}
package dbx

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"

	"github.com/ml444/gkit/log"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type (
	TxHandler  func(tx *gorm.DB) error
	TxCallback func() (model interface{}, execute func(scope *Scope) error)
)

func recoverTxPanic(err *error) {
	if r := recover(); r != nil {
		log.Errorf("Transaction PANIC: %v", r)
		log.Errorf("Stack trace:\n%s", string(debug.Stack()))
		*err = fmt.Errorf("Transaction PANIC: %v", r)
	}
}

func TxGo(ctx context.Context, db *gorm.DB, executes ...TxHandler) (err error) {
	defer recoverTxPanic(&err)
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
	defer recoverTxPanic(&err)

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

// ScopeTxGoWithT runs callbacks with Scope built from repository T (soft-delete filter, etc.).
func ScopeTxGoWithT(ctx context.Context, repo *T, callbacks ...TxCallback) (err error) {
	defer recoverTxPanic(&err)

	return repo.getDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, callback := range callbacks {
			model, exec := callback()
			if model == nil {
				model = repo.getModel()
			}
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

func RunTxItems(ctx context.Context, db *gorm.DB, items ...TxItem) (err error) {
	if len(items) == 0 {
		return nil
	}
	defer recoverTxPanic(&err)

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

// RunTxItemsWithT runs transaction items using Scope from repository T.
func RunTxItemsWithT(ctx context.Context, repo *T, items ...ScopeTxItem) (err error) {
	if len(items) == 0 {
		return nil
	}
	defer recoverTxPanic(&err)

	return repo.getDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			if err := item.Preload(repo, tx); err != nil {
				return err
			}
		}
		for _, item := range items {
			if err := item.Execute(repo, tx); err != nil {
				return err
			}
		}
		return nil
	})
}

// ScopeTxItem runs preload/execute through Scope (soft-delete, model binding).
type ScopeTxItem interface {
	Preload(repo *T, tx *gorm.DB) error
	Execute(repo *T, tx *gorm.DB) error
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

// ScopeInsertItem inserts models within a Scope (respects soft-delete model binding).
type ScopeInsertItem struct {
	Models any
}

func (i *ScopeInsertItem) Preload(repo *T, tx *gorm.DB) error {
	return nil
}

func (i *ScopeInsertItem) Execute(repo *T, tx *gorm.DB) error {
	if i.Models == nil {
		return nil
	}
	return NewScope(tx, i.Models).Create(i.Models)
}

type UpdateItem struct {
	Model   any
	Where   map[string]any
	Updates map[string]any
}

func (i *UpdateItem) Preload(tx *gorm.DB) error {
	if err := tx.Model(i.Model).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where(i.Where).
		First(i.Model).Error; err != nil {
		return err
	}
	return nil
}
func (i *UpdateItem) Execute(tx *gorm.DB) error {
	if i.Model == nil || i.Where == nil {
		return nil
	}
	if len(i.Updates) == 0 {
		return tx.Model(i.Model).Where(i.Where).Updates(i.Model).Error
	}
	return tx.Model(i.Model).Where(i.Where).Updates(i.Updates).Error
}

// ScopeUpdateItem updates via Scope with optional encryption through repository T.
type ScopeUpdateItem struct {
	Model   any
	Where   map[string]any
	Updates map[string]any
}

func (i *ScopeUpdateItem) Preload(repo *T, tx *gorm.DB) error {
	scope := NewScope(tx, i.Model)
	if len(i.Where) > 0 {
		if err := repo.CheckAndCrypto(i.Where, cipherKindEncrypt, false); err != nil {
			return err
		}
		scope = scope.Where(i.Where)
	}
	return scope.Clauses(clause.Locking{Strength: "UPDATE"}).First(i.Model)
}

func (i *ScopeUpdateItem) Execute(repo *T, tx *gorm.DB) error {
	if i.Model == nil || i.Where == nil {
		return nil
	}
	scope := NewScope(tx, i.Model).Where(i.Where)
	updates := i.Updates
	if updates == nil {
		updates = map[string]any{}
	}
	if err := repo.CheckAndCrypto(updates, cipherKindEncrypt, false); err != nil {
		return err
	}
	if len(i.Updates) == 0 {
		if err := repo.CheckAndCrypto(i.Model, cipherKindEncrypt, false); err != nil {
			return err
		}
		return scope.Update(i.Model)
	}
	return scope.Update(updates)
}

type SaveItem struct {
	Model any
	Where map[string]any
}

// Preload loads the model with row lock; record not found does not abort the transaction.
func (i *SaveItem) Preload(tx *gorm.DB) error {
	if i.Model == nil {
		return fmt.Errorf("model is nil")
	}
	if err := tx.Model(i.Model).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where(i.Where).
		First(i.Model).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
	}
	return nil
}
func (i *SaveItem) Execute(tx *gorm.DB) error {
	if i.Model == nil {
		return nil
	}
	return tx.Model(i.Model).Save(i.Model).Error
}

// ScopeSaveItem saves via Scope with soft-delete and encryption support.
type ScopeSaveItem struct {
	Model any
	Where map[string]any
}

func (i *ScopeSaveItem) Preload(repo *T, tx *gorm.DB) error {
	if i.Model == nil {
		return fmt.Errorf("model is nil")
	}
	scope := NewScope(tx, i.Model)
	if len(i.Where) > 0 {
		if err := repo.CheckAndCrypto(i.Where, cipherKindEncrypt, false); err != nil {
			return err
		}
		scope = scope.Where(i.Where)
	}
	if err := scope.Clauses(clause.Locking{Strength: "UPDATE"}).First(i.Model); err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
	}
	return nil
}

func (i *ScopeSaveItem) Execute(repo *T, tx *gorm.DB) error {
	if i.Model == nil {
		return nil
	}
	if err := repo.CheckAndCrypto(i.Model, cipherKindEncrypt, false); err != nil {
		return err
	}
	return NewScope(tx, i.Model).Save(i.Model)
}

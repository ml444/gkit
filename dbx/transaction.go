package dbx

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"

	"github.com/ml444/gkit/log"
)

type TxCallback func() (model any, execute func(scope *Scope) error)

func recoverTxPanic(err *error) {
	if r := recover(); r != nil {
		log.Errorf("Transaction PANIC: %v", r)
		log.Errorf("Stack trace:\n%s", string(debug.Stack()))
		*err = fmt.Errorf("Transaction PANIC: %v", r)
	}
}

func TxGo(ctx context.Context, conn Conn, executes ...func(d Driver) error) (err error) {
	defer recoverTxPanic(&err)
	return conn.Driver(ctx).Transaction(ctx, func(d Driver) error {
		for _, execute := range executes {
			if err := execute(d); err != nil {
				return err
			}
		}
		return nil
	})
}

func ScopeTxGo(ctx context.Context, conn Conn, callbacks ...TxCallback) (err error) {
	defer recoverTxPanic(&err)

	return conn.Driver(ctx).Transaction(ctx, func(d Driver) error {
		for _, callback := range callbacks {
			model, exec := callback()
			scope := newScopeWithDriver(wrapTxConn(conn, d), d, ctx, model)
			if err := exec(scope); err != nil {
				return err
			}
		}
		return nil
	})
}

func ScopeTxGoWithT(ctx context.Context, repo *T, callbacks ...TxCallback) (err error) {
	defer recoverTxPanic(&err)

	conn := repo.getConn()
	return conn.Driver(ctx).Transaction(ctx, func(d Driver) error {
		for _, callback := range callbacks {
			model, exec := callback()
			if model == nil {
				model = repo.getModel()
			}
			scope := newScopeWithDriver(wrapTxConn(conn, d), d, ctx, model)
			if err := exec(scope); err != nil {
				return err
			}
		}
		return nil
	})
}

func TxCreateMultiModels(ctx context.Context, conn Conn, models ...any) error {
	return conn.Driver(ctx).Transaction(ctx, func(d Driver) error {
		for _, m := range models {
			b := newQueryBuilder(m)
			if _, err := d.Create(ctx, b, m); err != nil {
				return err
			}
		}
		return nil
	})
}

type TxItem interface {
	Preload(d Driver) error
	Execute(d Driver) error
}

func RunTxItems(ctx context.Context, conn Conn, items ...TxItem) (err error) {
	if len(items) == 0 {
		return nil
	}
	defer recoverTxPanic(&err)

	return conn.Driver(ctx).Transaction(ctx, func(d Driver) error {
		for _, item := range items {
			if err := item.Preload(d); err != nil {
				return err
			}
		}
		for _, item := range items {
			if err := item.Execute(d); err != nil {
				return err
			}
		}
		return nil
	})
}

type ScopeTxItem interface {
	Preload(repo *T, d Driver) error
	Execute(repo *T, d Driver) error
}

func RunTxItemsWithT(ctx context.Context, repo *T, items ...ScopeTxItem) (err error) {
	if len(items) == 0 {
		return nil
	}
	defer recoverTxPanic(&err)

	conn := repo.getConn()
	return conn.Driver(ctx).Transaction(ctx, func(d Driver) error {
		for _, item := range items {
			if err := item.Preload(repo, d); err != nil {
				return err
			}
		}
		for _, item := range items {
			if err := item.Execute(repo, d); err != nil {
				return err
			}
		}
		return nil
	})
}

func NewInsertItem(models any) *InsertItem {
	return &InsertItem{Models: models}
}

type InsertItem struct {
	Models any
}

func (i *InsertItem) Preload(d Driver) error { return nil }

func (i *InsertItem) Execute(d Driver) error {
	if i.Models == nil {
		return nil
	}
	b := newQueryBuilder(i.Models)
	_, err := d.Create(context.Background(), b, i.Models)
	return err
}

type ScopeInsertItem struct {
	Models any
}

func (i *ScopeInsertItem) Preload(repo *T, d Driver) error { return nil }

func (i *ScopeInsertItem) Execute(repo *T, d Driver) error {
	if i.Models == nil {
		return nil
	}
	scope := newScopeWithDriver(wrapTxConn(repo.getConn(), d), d, context.Background(), i.Models)
	return scope.Create(i.Models)
}

type UpdateItem struct {
	Model   any
	Where   map[string]any
	Updates map[string]any
}

func (i *UpdateItem) Preload(d Driver) error {
	scope := newScopeWithDriver(nil, d, context.Background(), i.Model).SetForUpdate()
	if len(i.Where) > 0 {
		scope = scope.Where(i.Where)
	}
	return scope.First(i.Model)
}

func (i *UpdateItem) Execute(d Driver) error {
	if i.Model == nil || i.Where == nil {
		return nil
	}
	scope := newScopeWithDriver(nil, d, context.Background(), i.Model).Where(i.Where)
	if len(i.Updates) == 0 {
		return scope.Update(i.Model)
	}
	return scope.Update(i.Updates)
}

type ScopeUpdateItem struct {
	Model   any
	Where   map[string]any
	Updates map[string]any
}

func (i *ScopeUpdateItem) Preload(repo *T, d Driver) error {
	scope := newScopeWithDriver(wrapTxConn(repo.getConn(), d), d, context.Background(), i.Model).SetForUpdate()
	if len(i.Where) > 0 {
		if err := repo.CheckAndCrypto(i.Where, cipherKindEncrypt, false); err != nil {
			return err
		}
		scope = scope.Where(i.Where)
	}
	return scope.First(i.Model)
}

func (i *ScopeUpdateItem) Execute(repo *T, d Driver) error {
	if i.Model == nil || i.Where == nil {
		return nil
	}
	scope := newScopeWithDriver(wrapTxConn(repo.getConn(), d), d, context.Background(), i.Model).Where(i.Where)
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

func (i *SaveItem) Preload(d Driver) error {
	if i.Model == nil {
		return fmt.Errorf("model is nil")
	}
	scope := newScopeWithDriver(nil, d, context.Background(), i.Model).SetForUpdate()
	if len(i.Where) > 0 {
		scope = scope.Where(i.Where)
	}
	if err := scope.First(i.Model); err != nil {
		if !errors.Is(err, ErrRecordNotFound) {
			return err
		}
	}
	return nil
}

func (i *SaveItem) Execute(d Driver) error {
	if i.Model == nil {
		return nil
	}
	scope := newScopeWithDriver(nil, d, context.Background(), i.Model)
	return scope.Save(i.Model)
}

type ScopeSaveItem struct {
	Model any
	Where map[string]any
}

func (i *ScopeSaveItem) Preload(repo *T, d Driver) error {
	if i.Model == nil {
		return fmt.Errorf("model is nil")
	}
	scope := newScopeWithDriver(wrapTxConn(repo.getConn(), d), d, context.Background(), i.Model).SetForUpdate()
	if len(i.Where) > 0 {
		if err := repo.CheckAndCrypto(i.Where, cipherKindEncrypt, false); err != nil {
			return err
		}
		scope = scope.Where(i.Where)
	}
	if err := scope.First(i.Model); err != nil {
		if !errors.Is(err, ErrRecordNotFound) {
			return err
		}
	}
	return nil
}

func (i *ScopeSaveItem) Execute(repo *T, d Driver) error {
	if i.Model == nil {
		return nil
	}
	if err := repo.CheckAndCrypto(i.Model, cipherKindEncrypt, false); err != nil {
		return err
	}
	scope := newScopeWithDriver(wrapTxConn(repo.getConn(), d), d, context.Background(), i.Model)
	return scope.Save(i.Model)
}

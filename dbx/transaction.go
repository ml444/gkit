package dbx

import "gorm.io/gorm"

type TxHandler func(tx *gorm.DB) error
type TxCallback func() (model interface{}, execute func(scope *Scope) error)

func TxGo(db *gorm.DB, executes ...TxHandler) error {
	tx := db
	if _, ok := tx.Statement.ConnPool.(gorm.Tx); !ok {
		tx = tx.Begin()
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, execute := range executes {
		if err := execute(tx); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}
func ScopeTxGo(db *gorm.DB, callbacks ...TxCallback) error {
	tx := db
	if _, ok := tx.Statement.ConnPool.(gorm.Tx); !ok {
		tx = tx.Begin()
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, callback := range callbacks {
		model, exec := callback()
		scope := NewScope(tx, model)
		if err := exec(scope); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

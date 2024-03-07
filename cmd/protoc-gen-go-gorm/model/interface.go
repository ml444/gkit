package model

type IModel interface {
	ToORM() ITModel
}
type ITModel interface {
	ToSource() IModel
}

package user

type User struct {
	Id        uint64 `gorm:"primaryKey"`
	Name      string
	Age       uint32
	CreatedAt uint32
	UpdatedAt uint32
	DeletedAt uint32
}

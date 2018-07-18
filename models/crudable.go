package models

// CRUDable defines the crud methods
type CRUDable interface {
	Create(*User) error
	ReadOne(int64) error
	ReadAll(*User) (interface{}, error)
	Update(int64) error
	Delete() error
}
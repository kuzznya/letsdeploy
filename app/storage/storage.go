package storage

import "github.com/jmoiron/sqlx"

type Storage struct {
	ProjectRepository ProjectRepository
}

type CrudRepository[T any, ID any] interface {
	CreateNew(record T) (ID, error)
	FindByID(id ID) (T, error)
	Update(id ID, record T) error
	Delete(id ID) error
}

func New(db *sqlx.DB) *Storage {
	return &Storage{
		ProjectRepository: &projectRepositoryImpl{db: db},
	}
}

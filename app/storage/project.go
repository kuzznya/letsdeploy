package storage

import "github.com/jmoiron/sqlx"

type ProjectDBModel struct {
	Id   int    `db:"id"`
	Name string `db:"name"`
}

type ProjectRepository interface {
	CrudRepository[ProjectDBModel, int]
	FindUserProjects(username string)
}

type projectRepositoryImpl struct {
	db *sqlx.DB
}

func (p *projectRepositoryImpl) CreateNew(project ProjectDBModel) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (p *projectRepositoryImpl) FindByID(id int) (ProjectDBModel, error) {
	//TODO implement me
	panic("implement me")
}

func (p *projectRepositoryImpl) Update(id int, record ProjectDBModel) error {
	//TODO implement me
	panic("implement me")
}

func (p *projectRepositoryImpl) Delete(id int) error {
	//TODO implement me
	panic("implement me")
}

func (p *projectRepositoryImpl) FindUserProjects(username string) {
	//TODO implement me
	panic("implement me")
}

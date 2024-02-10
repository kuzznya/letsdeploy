package storage

import (
	"database/sql"
	"github.com/kuzznya/letsdeploy/app/apperrors"
	"github.com/pkg/errors"
)

type ManagedServiceEntity struct {
	Id        int    `db:"id"`
	ProjectId string `db:"project_id"`
	Name      string `db:"name"`
	Type      string `db:"type"`
}

type ManagedServiceRepository interface {
	CrudRepository[ManagedServiceEntity, int]
	FindAll(limit int, offset int) ([]ManagedServiceEntity, error)
	FindByProjectId(projectId string) ([]ManagedServiceEntity, error)
	ExistsByNameAndProjectId(name string, projectId string) (bool, error)
}

type managedServiceRepositoryImpl struct {
	db QueryExecDB
}

func (r managedServiceRepositoryImpl) CreateNew(entity ManagedServiceEntity) (int, error) {
	var id int
	err := r.db.Get(&id,
		"INSERT INTO managed_service (project_id, name, type) VALUES ($1, $2, $3) RETURNING id",
		entity.ProjectId, entity.Name, entity.Type)
	if err != nil {
		return 0, errors.Wrap(err, "cannot save new managed service")
	}
	return id, nil
}

func (r managedServiceRepositoryImpl) ExistsByID(id int) (bool, error) {
	var exists bool
	err := r.db.Get(&exists, "SELECT exists(SELECT * FROM managed_service WHERE id = $1)", id)
	if err != nil {
		return false, errors.Wrap(err, "cannot check if managed service exists")
	}
	return exists, nil
}

func (r managedServiceRepositoryImpl) FindByID(id int) (*ManagedServiceEntity, error) {
	var entity ManagedServiceEntity
	err := r.db.Get(&entity, "SELECT * FROM managed_service WHERE id = $1", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.NotFound("Project not found")
	} else if err != nil {
		return nil, errors.Wrap(err, "cannot find managed service by id")
	}
	return &entity, nil
}

func (r managedServiceRepositoryImpl) Update(entity ManagedServiceEntity) error {
	_, err := r.db.Exec("UPDATE managed_service SET name = $1, type = $2 WHERE id = $3",
		entity.Name, entity.Type, entity.Id)
	if err != nil {
		return errors.Wrap(err, "cannot update managed service")
	}
	return nil
}

func (r managedServiceRepositoryImpl) Delete(id int) error {
	_, err := r.db.Exec("DELETE FROM managed_service WHERE id = $1", id)
	if err != nil {
		return errors.Wrap(err, "failed to delete managed service")
	}
	return nil
}

func (r managedServiceRepositoryImpl) FindAll(limit int, offset int) ([]ManagedServiceEntity, error) {
	entities := []ManagedServiceEntity{}
	err := r.db.Select(&entities, "SELECT * FROM managed_service ORDER BY id LIMIT $1 OFFSET $2", limit, offset)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve managed services")
	}
	return entities, nil
}

func (r managedServiceRepositoryImpl) FindByProjectId(projectId string) ([]ManagedServiceEntity, error) {
	entities := []ManagedServiceEntity{}
	err := r.db.Select(&entities, "SELECT * FROM managed_service WHERE project_id = $1 ORDER BY name", projectId)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve managed services of a project")
	}
	return entities, nil
}

func (r managedServiceRepositoryImpl) ExistsByNameAndProjectId(name string, projectId string) (bool, error) {
	var exists bool
	err := r.db.Get(&exists,
		"SELECT exists(SELECT * FROM service WHERE project_id = $1 AND name = $2)",
		projectId, name)
	if err != nil {
		return false, errors.Wrap(err, "failed to check if project exists")
	}
	return exists, nil
}

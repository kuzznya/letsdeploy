package storage

import (
	"database/sql"
	"github.com/kuzznya/letsdeploy/app/apperrors"
	"github.com/pkg/errors"
)

type ServiceEntity struct {
	Id        int    `db:"id"`
	ProjectId string `db:"project_id"`
	Name      string `db:"name"`
	Image     string `db:"image"`
	Port      int    `db:"port"`
}

type ServiceRepository interface {
	CrudRepository[ServiceEntity, int]
	FindAll(limit int, offset int) ([]ServiceEntity, error)
	FindByProjectId(projectId string) ([]ServiceEntity, error)
	ExistsByNameAndProjectId(name string, projectId string) (bool, error)
}

type serviceRepositoryImpl struct {
	db QueryExecDB
}

func (r *serviceRepositoryImpl) CreateNew(service ServiceEntity) (int, error) {
	var id int
	err := r.db.Get(&id, "INSERT INTO service (project_id, name, image, port) VALUES ($1, $2, $3, $4) RETURNING id",
		service.ProjectId, service.Name, service.Image, service.Port)
	if err != nil {
		return 0, errors.Wrap(err, "cannot save new service")
	}
	return id, nil
}

func (r *serviceRepositoryImpl) ExistsByID(id int) (bool, error) {
	var exists bool
	err := r.db.Get(&exists, "SELECT exists(SELECT * FROM service WHERE id = $1)", id)
	if err != nil {
		return false, errors.Wrap(err, "cannot check if service exists")
	}
	return exists, nil
}

func (r *serviceRepositoryImpl) FindByID(id int) (*ServiceEntity, error) {
	var service ServiceEntity
	err := r.db.Get(&service, "SELECT * FROM service WHERE id = $1", id)
	if err == sql.ErrNoRows {
		return nil, apperrors.NotFound("Project not found")
	} else if err != nil {
		return nil, errors.Wrap(err, "cannot find service by id")
	}
	return &service, nil
}

func (r *serviceRepositoryImpl) Update(service ServiceEntity) error {
	_, err := r.db.Exec("UPDATE service SET name = $1, image = $2, port = $3 WHERE id = $4",
		service.Name, service.Image, service.Port, service.Id)
	if err != nil {
		return errors.Wrap(err, "failed to update service")
	}
	return nil
}

func (r *serviceRepositoryImpl) Delete(id int) error {
	_, err := r.db.Exec("DELETE FROM service WHERE id = $1", id)
	if err != nil {
		return errors.Wrap(err, "failed to delete service")
	}
	return nil
}

func (r *serviceRepositoryImpl) FindAll(limit int, offset int) ([]ServiceEntity, error) {
	services := []ServiceEntity{}
	err := r.db.Select(&services, "SELECT * FROM service LIMIT $1 OFFSET $2", limit, offset)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve services")
	}
	return services, nil
}

func (r *serviceRepositoryImpl) FindByProjectId(projectId string) ([]ServiceEntity, error) {
	services := []ServiceEntity{}
	err := r.db.Select(&services, "SELECT * FROM service WHERE project_id = $1", projectId)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve project services")
	}
	return services, nil
}

func (r *serviceRepositoryImpl) ExistsByNameAndProjectId(name string, projectId string) (bool, error) {
	var exists bool
	err := r.db.Get(&exists,
		"SELECT exists(SELECT * FROM service WHERE project_id = $1 AND name = $2)",
		projectId, name)
	if err != nil {
		return false, errors.Wrap(err, "failed to check if project exists")
	}
	return exists, nil
}

package storage

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"github.com/kuzznya/letsdeploy/app/apperrors"
	"github.com/pkg/errors"
)

type ServiceEntity struct {
	Id              int            `db:"id"`
	ProjectId       string         `db:"project_id"`
	Name            string         `db:"name"`
	Image           string         `db:"image"`
	Port            int            `db:"port"`
	PublicApiPrefix sql.NullString `db:"public_api_prefix"`
	EnvVars         EnvVars        `db:"env_vars"`
}

type EnvVarEntity struct {
	Name   string  `json:"name"`
	Value  *string `json:"value"`
	Secret *string `json:"secret"`
}

type EnvVars []EnvVarEntity

func (ev *EnvVars) Value() (driver.Value, error) {
	return json.Marshal(ev)
}

func (ev *EnvVars) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &ev)
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

func (r serviceRepositoryImpl) CreateNew(service ServiceEntity) (int, error) {
	var id int
	err := r.db.Get(&id,
		`INSERT INTO service (project_id, name, image, port, public_api_prefix, env_vars) 
		VALUES ($1, $2, $3, $4, $5, $6) 
		RETURNING id`,
		service.ProjectId, service.Name, service.Image, service.Port, service.PublicApiPrefix, &service.EnvVars)
	if err != nil {
		return 0, errors.Wrap(err, "cannot save new service")
	}
	return id, nil
}

func (r serviceRepositoryImpl) ExistsByID(id int) (bool, error) {
	var exists bool
	err := r.db.Get(&exists, "SELECT exists(SELECT * FROM service WHERE id = $1)", id)
	if err != nil {
		return false, errors.Wrap(err, "cannot check if service exists")
	}
	return exists, nil
}

func (r serviceRepositoryImpl) FindByID(id int) (*ServiceEntity, error) {
	var service ServiceEntity
	err := r.db.Get(&service, "SELECT * FROM service WHERE id = $1", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.NotFound("Project not found")
	} else if err != nil {
		return nil, errors.Wrap(err, "cannot find service by id")
	}
	return &service, nil
}

func (r serviceRepositoryImpl) Update(service ServiceEntity) error {
	_, err := r.db.Exec("UPDATE service SET name = $1, image = $2, port = $3, public_api_prefix = $4, env_vars = $5 WHERE id = $6",
		service.Name, service.Image, service.Port, service.PublicApiPrefix, &service.EnvVars, service.Id)
	if err != nil {
		return errors.Wrap(err, "failed to update service")
	}
	return nil
}

func (r serviceRepositoryImpl) Delete(id int) error {
	_, err := r.db.Exec("DELETE FROM service WHERE id = $1", id)
	if err != nil {
		return errors.Wrap(err, "failed to delete service")
	}
	return nil
}

func (r serviceRepositoryImpl) FindAll(limit int, offset int) ([]ServiceEntity, error) {
	services := []ServiceEntity{}
	err := r.db.Select(&services, "SELECT * FROM service ORDER BY id LIMIT $1 OFFSET $2", limit, offset)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve services")
	}
	return services, nil
}

func (r serviceRepositoryImpl) FindByProjectId(projectId string) ([]ServiceEntity, error) {
	services := []ServiceEntity{}
	err := r.db.Select(&services, "SELECT * FROM service WHERE project_id = $1 ORDER BY name", projectId)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve project services")
	}
	return services, nil
}

func (r serviceRepositoryImpl) ExistsByNameAndProjectId(name string, projectId string) (bool, error) {
	var exists bool
	err := r.db.Get(&exists,
		"SELECT exists(SELECT * FROM service WHERE project_id = $1 AND name = $2)",
		projectId, name)
	if err != nil {
		return false, errors.Wrap(err, "failed to check if project exists")
	}
	return exists, nil
}

package storage

import (
	"database/sql"
	"github.com/pkg/errors"
)

type EnvVarEntity struct {
	Id        int            `db:"id"`
	ServiceId int            `db:"service_id"`
	Name      string         `db:"name"`
	Value     sql.NullString `db:"value"`
	Secret    sql.NullString `db:"secret"`
}

type EnvVarRepository interface {
	FindByServiceId(id int) ([]EnvVarEntity, error)
	CreateOrUpdate(entity EnvVarEntity) (*EnvVarEntity, error)
	DeleteByServiceIdAndName(serviceId int, name string) error
}

type envVarRepositoryImpl struct {
	db QueryExecDB
}

func (e envVarRepositoryImpl) FindByServiceId(id int) ([]EnvVarEntity, error) {
	entities := []EnvVarEntity{}
	err := e.db.Select(&entities, "SELECT * FROM env_var WHERE service_id = $1", id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get service env vars")
	}
	return entities, nil
}

func (e envVarRepositoryImpl) CreateOrUpdate(entity EnvVarEntity) (*EnvVarEntity, error) {
	saved := EnvVarEntity{}
	query := `INSERT INTO env_var (service_id, name, value, secret_id) 
			  VALUES ($1, $2, $3, (SELECT id FROM secret WHERE name = $4))
			  ON CONFLICT (service_id, name) DO UPDATE 
			      SET name = $2, value = $3, secret_id = (SELECT id FROM secret WHERE name = $4) 
			  RETURNING *`
	err := e.db.Get(&saved, query, entity.ServiceId, entity.Name, entity.Value, entity.Secret)
	if err != nil {
		return nil, err
	}
	return &saved, nil
}

func (e envVarRepositoryImpl) DeleteByName(name string) error {
	_, err := e.db.Exec("DELETE FROM env_var WHERE name = $1", name)
	if err != nil {
		return errors.Wrap(err, "failed to delete env var")
	}
	return nil
}

func (e envVarRepositoryImpl) DeleteByServiceIdAndName(serviceId int, name string) error {
	_, err := e.db.Exec("DELETE FROM env_var WHERE service_id = $1 AND name = $2", serviceId, name)
	if err != nil {
		return errors.Wrap(err, "failed to delete env var")
	}
	return nil
}

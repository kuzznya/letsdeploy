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
	Secret    sql.NullString
}

type EnvVarRepository interface {
	FindByServiceId(id int) ([]EnvVarEntity, error)
	CreateOrUpdate(entity EnvVarEntity) (*EnvVarEntity, error)
	CreateOrUpdateAll(vars []EnvVarEntity) ([]EnvVarEntity, error)
	DeleteByServiceIdAndName(serviceId int, name string) error
}

type envVarRepositoryImpl struct {
	db QueryExecDB
}

var _ EnvVarRepository = (*envVarRepositoryImpl)(nil)

func (e envVarRepositoryImpl) FindByServiceId(id int) ([]EnvVarEntity, error) {
	entities := []EnvVarEntity{}
	query := `
SELECT e.id, e.service_id, e.name, e.value, s.name as secret 
FROM env_var e
LEFT JOIN secret s ON s.id = e.secret_id
WHERE service_id = $1
`
	err := e.db.Select(&entities, query, id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get service env vars")
	}
	return entities, nil
}

func (e envVarRepositoryImpl) CreateOrUpdate(entity EnvVarEntity) (*EnvVarEntity, error) {
	saved := EnvVarEntity{}
	query := `
INSERT INTO env_var (service_id, name, value, secret_id) 
VALUES ($1, $2, $3, (SELECT id FROM secret WHERE name = $4))
ON CONFLICT (service_id, name) DO UPDATE 
    SET name = $2, value = $3, secret_id = (SELECT id FROM secret WHERE name = $4) 
RETURNING  env_var.id, env_var.service_id, env_var.name, env_var.value, 
    (SELECT name FROM secret WHERE id = env_var.secret_id) as secret`
	err := e.db.Get(&saved, query, entity.ServiceId, entity.Name, entity.Value, entity.Secret)
	if err != nil {
		return nil, err
	}
	return &saved, nil
}

func (e envVarRepositoryImpl) CreateOrUpdateAll(vars []EnvVarEntity) ([]EnvVarEntity, error) {

	query := `
INSERT INTO env_var (service_id, name, value, secret_id)
VALUES (:service_id, :name, :value, (SELECT id FROM secret WHERE name = :secret))
ON CONFLICT (service_id, name) DO UPDATE 
    SET value = excluded.value, secret_id = excluded.secret_id 
RETURNING env_var.id, env_var.service_id, env_var.name, env_var.value, 
    (SELECT name FROM secret WHERE id = env_var.secret_id) as secret`
	rows, err := e.db.NamedQuery(query, vars)
	if err != nil {
		return nil, errors.Wrap(err, "failed to save env vars")
	}

	saved := make([]EnvVarEntity, 0)
	next := rows.Next()
	for next {
		var v EnvVarEntity
		if err := rows.StructScan(&v); err != nil {
			return nil, errors.Wrap(err, "failed to map SQL row to env var entity")
		}
		saved = append(saved, v)
		next = rows.Next()
	}
	return saved, nil
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

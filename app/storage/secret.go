package storage

import (
	"database/sql"
	"fmt"
	"github.com/kuzznya/letsdeploy/app/apperrors"
	"github.com/pkg/errors"
)

type SecretEntity struct {
	Id               int    `db:"id"`
	ProjectId        string `db:"project_id"`
	Name             string `db:"name"`
	Value            string `db:"value"`
	ManagedServiceId *int   `db:"managed_service_id"`
}

type SecretRepository interface {
	FindByProjectId(id string) ([]SecretEntity, error)
	CreateNew(secret SecretEntity) error
	ExistsByProjectIdAndName(id string, name string) (bool, error)
	FindByProjectIdAndName(id string, name string) (*SecretEntity, error)
	DeleteByProjectIdAndName(id string, name string) error
}

type secretRepositoryImpl struct {
	db QueryExecDB
}

func (s secretRepositoryImpl) FindByProjectId(id string) ([]SecretEntity, error) {
	secrets := []SecretEntity{}
	err := s.db.Select(&secrets, "SELECT * FROM secret WHERE project_id = $1", id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project secrets")
	}
	return secrets, nil
}

func (s secretRepositoryImpl) CreateNew(secret SecretEntity) error {
	_, err := s.db.Exec(
		"INSERT INTO secret (project_id, name, value, managed_service_id) VALUES ($1, $2, $3, $4)",
		secret.ProjectId, secret.Name, secret.Value, secret.ManagedServiceId)
	if err != nil {
		return errors.Wrap(err, "failed to create new secret")
	}
	return nil
}

func (s secretRepositoryImpl) ExistsByProjectIdAndName(id string, name string) (bool, error) {
	var exists bool
	err := s.db.Get(&exists, "SELECT exists(SELECT * FROM secret WHERE project_id = $1 AND name = $2)", id, name)
	if err != nil {
		return false, errors.Wrap(err, "failed to get secret by name")
	}
	return exists, nil
}

func (s secretRepositoryImpl) FindByProjectIdAndName(id string, name string) (*SecretEntity, error) {
	var secret SecretEntity
	err := s.db.Get(&secret, "SELECT * FROM secret WHERE project_id = $1 AND name = $2", id, name)
	if err == sql.ErrNoRows {
		return nil, apperrors.NotFound(fmt.Sprintf("SecretId with name %s not found", name))
	} else if err != nil {
		return nil, errors.Wrap(err, "failed to get secret by name")
	}
	return &secret, nil
}

func (s secretRepositoryImpl) DeleteByProjectIdAndName(id string, name string) error {
	_, err := s.db.Exec("DELETE FROM secret WHERE project_id = $1 AND name = $2", id, name)
	if err != nil {
		return errors.Wrap(err, "failed to delete secret")
	}
	return nil
}

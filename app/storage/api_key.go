package storage

import (
	"database/sql"
	"github.com/kuzznya/letsdeploy/app/apperrors"
	"github.com/pkg/errors"
)

type ApiKeyEntity struct {
	Id       string `db:"id"`
	Username string `db:"username"`
	Name     string `db:"name"`
}

type ApiKeyRepository interface {
	CrudRepository[ApiKeyEntity, string]
	GetByUsername(username string) ([]ApiKeyEntity, error)
}

type apiKeyRepositoryImpl struct {
	db QueryExecDB
}

func (r apiKeyRepositoryImpl) CreateNew(apiKey ApiKeyEntity) (string, error) {
	_, err := r.db.Exec("INSERT INTO api_key (id, name, username) VALUES ($1, $2, $3)",
		apiKey.Id, apiKey.Name, apiKey.Username)
	if err != nil {
		return "", errors.Wrap(err, "failed to create the new API key")
	}
	return apiKey.Id, nil
}

func (r apiKeyRepositoryImpl) ExistsByID(id string) (bool, error) {
	var exists bool
	err := r.db.Select(&exists, "SELECT exists(SELECT * FROM api_key WHERE id = $1)", id)
	if err != nil {
		return false, errors.Wrap(err, "failed to check if API key exists")
	}
	return exists, nil
}

func (r apiKeyRepositoryImpl) FindByID(id string) (*ApiKeyEntity, error) {
	var apiKey ApiKeyEntity
	err := r.db.Get(&apiKey, "SELECT * FROM api_key WHERE id = $1", id)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.NotFound("API key not found")
	}
	return &apiKey, nil
}

func (r apiKeyRepositoryImpl) Update(record ApiKeyEntity) error {
	panic("not implemented")
}

func (r apiKeyRepositoryImpl) Delete(id string) error {
	_, err := r.db.Exec("DELETE FROM api_key WHERE id = $1", id)
	if err != nil {
		return errors.Wrap(err, "failed to delete the API key")
	}
	return nil
}

func (r apiKeyRepositoryImpl) GetByUsername(username string) ([]ApiKeyEntity, error) {
	apiKeys := []ApiKeyEntity{}
	err := r.db.Select(&apiKeys, "SELECT * FROM api_key WHERE username = $1", username)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve user's API keys")
	}
	return apiKeys, nil
}

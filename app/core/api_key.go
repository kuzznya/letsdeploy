package core

import (
	"context"
	"github.com/kuzznya/letsdeploy/app/apperrors"
	"github.com/kuzznya/letsdeploy/app/middleware"
	"github.com/kuzznya/letsdeploy/app/storage"
	"github.com/kuzznya/letsdeploy/internal/openapi"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"math/rand"
)

const apiKeyPrefix = "ldp_"

type ApiKeys interface {
	GetApiKeys(auth middleware.Authentication) ([]openapi.ApiKey, error)
	GetUsernameByApiKey(apiKey string) (string, error)
	CreateApiKey(ctx context.Context, key openapi.ApiKey, auth middleware.Authentication) (*openapi.ApiKey, error)
	DeleteApiKey(ctx context.Context, apiKey string, auth middleware.Authentication) error
}

type apiKeysImpl struct {
	storage *storage.Storage
}

var _ ApiKeys = (*apiKeysImpl)(nil)

func InitApiKeys(storage *storage.Storage) ApiKeys {
	return &apiKeysImpl{storage: storage}
}

func (a apiKeysImpl) GetApiKeys(auth middleware.Authentication) ([]openapi.ApiKey, error) {
	entities, err := a.storage.ApiKeyRepository().GetByUsername(auth.Username)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get API keys")
	}
	keys := make([]openapi.ApiKey, len(entities))
	for i, entity := range entities {
		id := entity.Id
		keys[i] = openapi.ApiKey{
			Key:  &id,
			Name: entity.Name,
		}
	}
	return keys, nil
}

func (a apiKeysImpl) GetUsernameByApiKey(apiKey string) (string, error) {
	entity, err := a.storage.ApiKeyRepository().FindByID(apiKey)
	if err != nil {
		return "", errors.Wrap(err, "failed to get API key")
	}
	return entity.Username, nil
}

func (a apiKeysImpl) CreateApiKey(ctx context.Context, key openapi.ApiKey, auth middleware.Authentication) (*openapi.ApiKey, error) {
	keyId := generateApiKey()
	key.Key = &keyId
	entity := storage.ApiKeyEntity{
		Id:       keyId,
		Name:     key.Name,
		Username: auth.Username,
	}
	err := a.storage.ExecTx(ctx, func(s *storage.Storage) error {
		_, err := s.ApiKeyRepository().CreateNew(entity)
		return err
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create the API key")
	}
	log.Infof("Created API key for user %s", auth.Username)
	return &key, nil
}

func (a apiKeysImpl) DeleteApiKey(ctx context.Context, apiKey string, auth middleware.Authentication) error {
	key, err := a.storage.ApiKeyRepository().FindByID(apiKey)
	if apperrors.IsNotFound(err) {
		return nil
	} else if err != nil {
		return errors.Wrap(err, "failed to get API key")
	}
	if key.Username != auth.Username {
		return apperrors.NotFound("API key not found")
	}
	err = a.storage.ExecTx(ctx, func(s *storage.Storage) error {
		return s.ApiKeyRepository().Delete(apiKey)
	})
	if err != nil {
		return errors.Wrap(err, "failed to delete API key")
	}
	log.Infof("Deleted API key for user %s", auth.Username)
	return nil
}

func generateApiKey() string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	tokenLen := 32
	b := make([]rune, tokenLen)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	apiKey := string(b)
	return apiKeyPrefix + apiKey
}

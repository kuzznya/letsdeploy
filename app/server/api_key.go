package server

import (
	"context"
	"github.com/kuzznya/letsdeploy/app/middleware"
	"github.com/kuzznya/letsdeploy/internal/openapi"
	"github.com/pkg/errors"
)

func (s Server) GetApiKeys(ctx context.Context, _ openapi.GetApiKeysRequestObject) (openapi.GetApiKeysResponseObject, error) {
	keys, err := s.core.ApiKeys.GetApiKeys(middleware.GetAuth(ctx))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get API keys")
	}
	return openapi.GetApiKeys200JSONResponse(keys), nil
}

func (s Server) CreateApiKey(ctx context.Context, request openapi.CreateApiKeyRequestObject) (openapi.CreateApiKeyResponseObject, error) {
	key, err := s.core.ApiKeys.CreateApiKey(ctx, *request.Body, middleware.GetAuth(ctx))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create API key")
	}
	return openapi.CreateApiKey200JSONResponse(*key), nil
}

func (s Server) DeleteApiKey(ctx context.Context, request openapi.DeleteApiKeyRequestObject) (openapi.DeleteApiKeyResponseObject, error) {
	err := s.core.ApiKeys.DeleteApiKey(ctx, request.Key, middleware.GetAuth(ctx))
	if err != nil {
		return nil, errors.Wrap(err, "failed to delete API key")
	}
	return openapi.DeleteApiKey200Response{}, nil
}

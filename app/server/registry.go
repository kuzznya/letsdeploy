package server

import (
	"context"
	"github.com/kuzznya/letsdeploy/app/middleware"
	"github.com/kuzznya/letsdeploy/internal/openapi"
	"github.com/pkg/errors"
)

func (s Server) GetProjectContainerRegistries(ctx context.Context, request openapi.GetProjectContainerRegistriesRequestObject) (openapi.GetProjectContainerRegistriesResponseObject, error) {
	regs, err := s.core.Registries.GetProjectContainerRegistries(request.Id, middleware.GetAuth(ctx))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project container registries")
	}
	return openapi.GetProjectContainerRegistries200JSONResponse(regs), nil
}

func (s Server) AddContainerRegistry(ctx context.Context, request openapi.AddContainerRegistryRequestObject) (openapi.AddContainerRegistryResponseObject, error) {
	reg, err := s.core.Registries.AddContainerRegistry(ctx, request.Id, *request.Body, middleware.GetAuth(ctx))
	if err != nil {
		return nil, errors.Wrap(err, "failed to add container registry")
	}
	return openapi.AddContainerRegistry200JSONResponse(reg), nil
}

func (s Server) DeleteContainerRegistry(ctx context.Context, request openapi.DeleteContainerRegistryRequestObject) (openapi.DeleteContainerRegistryResponseObject, error) {
	err := s.core.Registries.DeleteContainerRegistry(ctx, request.Id, request.RegistryId, middleware.GetAuth(ctx))
	if err != nil {
		return nil, errors.Wrap(err, "failed to delete container registry")
	}
	return openapi.DeleteContainerRegistry200Response{}, nil
}

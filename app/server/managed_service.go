package server

import (
	"context"
	"github.com/kuzznya/letsdeploy/app/middleware"
	"github.com/kuzznya/letsdeploy/internal/openapi"
	"github.com/pkg/errors"
)

func (s Server) CreateManagedService(ctx context.Context, request openapi.CreateManagedServiceRequestObject) (openapi.CreateManagedServiceResponseObject, error) {
	service, err := s.core.ManagedServices.CreateManagedService(ctx, *request.Body, middleware.GetAuth(ctx))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create managed service")
	}
	return openapi.CreateManagedService200JSONResponse(*service), nil
}

func (s Server) GetManagedService(ctx context.Context, request openapi.GetManagedServiceRequestObject) (openapi.GetManagedServiceResponseObject, error) {
	service, err := s.core.ManagedServices.GetManagedService(request.Id, middleware.GetAuth(ctx))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get managed service")
	}
	return openapi.GetManagedService200JSONResponse(*service), err
}

func (s Server) DeleteManagedService(ctx context.Context, request openapi.DeleteManagedServiceRequestObject) (openapi.DeleteManagedServiceResponseObject, error) {
	err := s.core.ManagedServices.DeleteManagedService(ctx, request.Id, middleware.GetAuth(ctx))
	if err != nil {
		return nil, err
	}
	return openapi.DeleteManagedService200Response{}, nil
}

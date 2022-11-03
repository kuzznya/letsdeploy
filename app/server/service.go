package server

import (
	"context"
	"github.com/kuzznya/letsdeploy/app/middleware"
	"github.com/kuzznya/letsdeploy/internal/openapi"
)

func (s Server) CreateService(ctx context.Context, request openapi.CreateServiceRequestObject) (openapi.CreateServiceResponseObject, error) {
	service, err := s.core.Services.CreateService(ctx, *request.Body, middleware.GetAuth(ctx).Username)
	if err != nil {
		return nil, err
	}
	return openapi.CreateService200JSONResponse(*service), nil
}

func (s Server) GetService(ctx context.Context, request openapi.GetServiceRequestObject) (openapi.GetServiceResponseObject, error) {
	service, err := s.core.Services.GetService(request.Id, middleware.GetAuth(ctx).Username)
	if err != nil {
		return nil, err
	}
	return openapi.GetService200JSONResponse(*service), nil
}

func (s Server) UpdateService(ctx context.Context, request openapi.UpdateServiceRequestObject) (openapi.UpdateServiceResponseObject, error) {
	request.Body.Id = &request.Id
	service, err := s.core.Services.UpdateService(ctx, *request.Body, middleware.GetAuth(ctx).Username)
	if err != nil {
		return nil, err
	}
	return openapi.UpdateService200JSONResponse(*service), nil
}

func (s Server) DeleteService(ctx context.Context, request openapi.DeleteServiceRequestObject) (openapi.DeleteServiceResponseObject, error) {
	err := s.core.Services.DeleteService(ctx, request.Id, middleware.GetAuth(ctx).Username)
	if err != nil {
		return nil, err
	}
	return openapi.DeleteService200Response{}, nil
}

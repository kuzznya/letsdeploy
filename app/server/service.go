package server

import (
	"context"
	"github.com/kuzznya/letsdeploy/app/middleware"
	"github.com/kuzznya/letsdeploy/internal/openapi"
	"github.com/pkg/errors"
)

func (s Server) CreateService(ctx context.Context, request openapi.CreateServiceRequestObject) (openapi.CreateServiceResponseObject, error) {
	service, err := s.core.Services.CreateService(ctx, *request.Body, middleware.GetAuth(ctx))
	if err != nil {
		return nil, err
	}
	return openapi.CreateService200JSONResponse(*service), nil
}

func (s Server) GetService(ctx context.Context, request openapi.GetServiceRequestObject) (openapi.GetServiceResponseObject, error) {
	service, err := s.core.Services.GetService(request.Id, middleware.GetAuth(ctx))
	if err != nil {
		return nil, err
	}
	return openapi.GetService200JSONResponse(*service), nil
}

func (s Server) UpdateService(ctx context.Context, request openapi.UpdateServiceRequestObject) (openapi.UpdateServiceResponseObject, error) {
	request.Body.Id = &request.Id
	service, err := s.core.Services.UpdateService(ctx, *request.Body, middleware.GetAuth(ctx))
	if err != nil {
		return nil, err
	}
	return openapi.UpdateService200JSONResponse(*service), nil
}

func (s Server) DeleteService(ctx context.Context, request openapi.DeleteServiceRequestObject) (openapi.DeleteServiceResponseObject, error) {
	err := s.core.Services.DeleteService(ctx, request.Id, middleware.GetAuth(ctx))
	if err != nil {
		return nil, err
	}
	return openapi.DeleteService200Response{}, nil
}

func (s Server) GetServiceStatus(ctx context.Context, request openapi.GetServiceStatusRequestObject) (openapi.GetServiceStatusResponseObject, error) {
	status, err := s.core.Services.GetServiceStatus(ctx, request.Id, middleware.GetAuth(ctx))
	if err != nil {
		return nil, err
	}
	return openapi.GetServiceStatus200JSONResponse(*status), nil
}

func (s Server) RestartService(ctx context.Context, request openapi.RestartServiceRequestObject) (openapi.RestartServiceResponseObject, error) {
	err := s.core.Services.RestartService(ctx, request.Id, middleware.GetAuth(ctx))
	if err != nil {
		return nil, err
	}
	return openapi.RestartService200Response{}, nil
}

func (s Server) GetServiceEnvVars(ctx context.Context, request openapi.GetServiceEnvVarsRequestObject) (openapi.GetServiceEnvVarsResponseObject, error) {
	envVars, err := s.core.Services.GetServiceEnvVars(request.Id, middleware.GetAuth(ctx))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get service env vars")
	}
	return openapi.GetServiceEnvVars200JSONResponse(envVars), nil
}

func (s Server) SetServiceEnvVar(ctx context.Context, request openapi.SetServiceEnvVarRequestObject) (openapi.SetServiceEnvVarResponseObject, error) {
	envVar, err := s.core.Services.SetServiceEnvVar(ctx, request.Id, *request.Body, middleware.GetAuth(ctx))
	if err != nil {
		return nil, errors.Wrap(err, "failed to set service env var")
	}
	return openapi.SetServiceEnvVar200JSONResponse(*envVar), nil
}

func (s Server) DeleteServiceEnvVar(ctx context.Context, request openapi.DeleteServiceEnvVarRequestObject) (openapi.DeleteServiceEnvVarResponseObject, error) {
	err := s.core.Services.DeleteServiceEnvVar(ctx, request.Id, request.Name, middleware.GetAuth(ctx))
	if err != nil {
		return nil, errors.Wrap(err, "failed to delete service env var")
	}
	return nil, err
}

package server

import (
	"context"
	"github.com/kuzznya/letsdeploy/app/middleware"
	"github.com/kuzznya/letsdeploy/internal/openapi"
	"github.com/pkg/errors"
)

func (s Server) GetProjects(ctx context.Context, _ openapi.GetProjectsRequestObject) (openapi.GetProjectsResponseObject, error) {
	projects, err := s.core.Projects.GetUserProjects(middleware.GetAuth(ctx))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user projects")
	}
	return openapi.GetProjects200JSONResponse(projects), nil
}

func (s Server) CreateProject(ctx context.Context, request openapi.CreateProjectRequestObject) (openapi.CreateProjectResponseObject, error) {
	project, err := s.core.Projects.CreateProject(ctx, *request.Body, middleware.GetAuth(ctx))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new project")
	}
	return openapi.CreateProject200JSONResponse(*project), nil
}

func (s Server) GetProject(ctx context.Context, request openapi.GetProjectRequestObject) (openapi.GetProjectResponseObject, error) {
	info, err := s.core.Projects.GetProjectInfo(request.Id, middleware.GetAuth(ctx))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project")
	}
	return openapi.GetProject200JSONResponse(*info), nil
}

func (s Server) DeleteProject(ctx context.Context, request openapi.DeleteProjectRequestObject) (openapi.DeleteProjectResponseObject, error) {
	err := s.core.Projects.DeleteProject(ctx, request.Id, middleware.GetAuth(ctx))
	if err != nil {
		return nil, errors.Wrap(err, "failed to delete project")
	}
	return openapi.DeleteProject200Response{}, nil
}

func (s Server) GetProjectParticipants(ctx context.Context, request openapi.GetProjectParticipantsRequestObject) (openapi.GetProjectParticipantsResponseObject, error) {
	participants, err := s.core.Projects.GetParticipants(request.Id, middleware.GetAuth(ctx))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project participants")
	}
	return (openapi.GetProjectParticipants200JSONResponse)(participants), nil
}

func (s Server) RemoveProjectParticipant(ctx context.Context, request openapi.RemoveProjectParticipantRequestObject) (openapi.RemoveProjectParticipantResponseObject, error) {
	err := s.core.Projects.RemoveParticipant(request.Id, request.Username, middleware.GetAuth(ctx))
	if err != nil {
		return nil, errors.Wrap(err, "failed to remove project participant")
	}
	return openapi.RemoveProjectParticipant200Response{}, nil
}

func (s Server) AddProjectParticipant(ctx context.Context, request openapi.AddProjectParticipantRequestObject) (openapi.AddProjectParticipantResponseObject, error) {
	err := s.core.Projects.AddParticipant(request.Id, request.Username, middleware.GetAuth(ctx))
	if err != nil {
		return nil, errors.Wrap(err, "failed to add project participant")
	}
	return openapi.AddProjectParticipant200Response{}, nil
}

func (s Server) JoinProject(ctx context.Context, request openapi.JoinProjectRequestObject) (openapi.JoinProjectResponseObject, error) {
	project, err := s.core.Projects.JoinProject(ctx, request.InviteCode, middleware.GetAuth(ctx))
	if err != nil {
		return nil, errors.Wrap(err, "failed to join project")
	}
	return openapi.JoinProject200JSONResponse(*project), nil
}

func (s Server) GetSecrets(ctx context.Context, request openapi.GetSecretsRequestObject) (openapi.GetSecretsResponseObject, error) {
	secrets, err := s.core.Projects.GetSecrets(request.Id, middleware.GetAuth(ctx))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project secrets")
	}
	return openapi.GetSecrets200JSONResponse(secrets), nil
}

func (s Server) CreateSecret(ctx context.Context, request openapi.CreateSecretRequestObject) (openapi.CreateSecretResponseObject, error) {
	secret, err := s.core.Projects.CreateSecret(
		ctx,
		request.Id,
		openapi.Secret{Name: request.Body.Name},
		request.Body.Value,
		middleware.GetAuth(ctx))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new secret")
	}
	return openapi.CreateSecret200JSONResponse(*secret), nil
}

func (s Server) DeleteSecret(ctx context.Context, request openapi.DeleteSecretRequestObject) (openapi.DeleteSecretResponseObject, error) {
	err := s.core.Projects.DeleteSecret(ctx, request.Id, request.Name, middleware.GetAuth(ctx))
	if err != nil {
		return nil, errors.Wrap(err, "failed to delete secret")
	}
	return openapi.DeleteSecret200Response{}, nil
}

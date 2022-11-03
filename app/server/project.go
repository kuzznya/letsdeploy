package server

import (
	"context"
	"github.com/kuzznya/letsdeploy/app/middleware"
	"github.com/kuzznya/letsdeploy/internal/openapi"
	"github.com/pkg/errors"
)

func (s Server) GetProjects(ctx context.Context, _ openapi.GetProjectsRequestObject) (openapi.GetProjectsResponseObject, error) {
	projects, err := s.core.Projects.GetUserProjects(middleware.GetAuth(ctx).Username)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user projects")
	}
	return openapi.GetProjects200JSONResponse(projects), nil
}

func (s Server) CreateProject(ctx context.Context, request openapi.CreateProjectRequestObject) (openapi.CreateProjectResponseObject, error) {
	project, err := s.core.Projects.CreateProject(ctx, *request.Body, middleware.GetAuth(ctx).Username)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new project")
	}
	return openapi.CreateProject200JSONResponse(*project), nil
}

func (s Server) GetProject(ctx context.Context, request openapi.GetProjectRequestObject) (openapi.GetProjectResponseObject, error) {
	info, err := s.core.Projects.GetProjectInfo(request.Id, middleware.GetAuth(ctx).Username)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project")
	}
	return openapi.GetProject200JSONResponse(*info), nil
}

func (s Server) DeleteProject(ctx context.Context, request openapi.DeleteProjectRequestObject) (openapi.DeleteProjectResponseObject, error) {
	err := s.core.Projects.DeleteProject(ctx, request.Id, middleware.GetAuth(ctx).Username)
	if err != nil {
		return nil, errors.Wrap(err, "failed to delete project")
	}
	return openapi.DeleteProject200Response{}, nil
}

func (s Server) GetProjectParticipants(ctx context.Context, request openapi.GetProjectParticipantsRequestObject) (openapi.GetProjectParticipantsResponseObject, error) {
	participants, err := s.core.Projects.GetParticipants(request.Id, middleware.GetAuth(ctx).Username)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project participants")
	}
	return (openapi.GetProjectParticipants200JSONResponse)(participants), nil
}

func (s Server) RemoveProjectParticipant(ctx context.Context, request openapi.RemoveProjectParticipantRequestObject) (openapi.RemoveProjectParticipantResponseObject, error) {
	err := s.core.Projects.RemoveParticipant(request.Id, request.Username, middleware.GetAuth(ctx).Username)
	if err != nil {
		return nil, errors.Wrap(err, "failed to remove project participant")
	}
	return openapi.RemoveProjectParticipant200Response{}, nil
}

func (s Server) AddProjectParticipant(ctx context.Context, request openapi.AddProjectParticipantRequestObject) (openapi.AddProjectParticipantResponseObject, error) {
	err := s.core.Projects.AddParticipant(request.Id, request.Username, middleware.GetAuth(ctx).Username)
	if err != nil {
		return nil, errors.Wrap(err, "failed to add project participant")
	}
	return openapi.AddProjectParticipant200Response{}, nil
}

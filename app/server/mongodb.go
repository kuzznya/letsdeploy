package server

import (
	"context"
	"github.com/kuzznya/letsdeploy/app/middleware"
	"github.com/kuzznya/letsdeploy/internal/openapi"
	"github.com/pkg/errors"
)

func (s Server) GetMongoDbUsers(ctx context.Context, request openapi.GetMongoDbUsersRequestObject) (openapi.GetMongoDbUsersResponseObject, error) {
	users, err := s.core.MongoDbMgmt.GetMongoDbUsers(ctx, request.Id, middleware.GetAuth(ctx))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get MongoDB users")
	}
	return openapi.GetMongoDbUsers200JSONResponse(users), nil
}

func (s Server) CreateMongoDbUser(ctx context.Context, request openapi.CreateMongoDbUserRequestObject) (openapi.CreateMongoDbUserResponseObject, error) {
	user, err := s.core.MongoDbMgmt.CreateMongoDbUser(ctx, request.Id, *request.Body, middleware.GetAuth(ctx))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create MongoDB user")
	}
	return openapi.CreateMongoDbUser200JSONResponse(user), nil
}

func (s Server) DeleteMongoDbUser(ctx context.Context, request openapi.DeleteMongoDbUserRequestObject) (openapi.DeleteMongoDbUserResponseObject, error) {
	err := s.core.MongoDbMgmt.DeleteMongoDbUser(ctx, request.Id, request.Username, middleware.GetAuth(ctx))
	if err != nil {
		return nil, errors.Wrap(err, "failed to delete MongoDB user")
	}
	return openapi.DeleteMongoDbUser200Response{}, nil
}

func (s Server) GetMongoDbUser(ctx context.Context, request openapi.GetMongoDbUserRequestObject) (openapi.GetMongoDbUserResponseObject, error) {
	user, err := s.core.MongoDbMgmt.GetMongoDbUser(ctx, request.Id, request.Username, middleware.GetAuth(ctx))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get MongoDB user")
	}
	return openapi.GetMongoDbUser200JSONResponse(user), nil
}

func (s Server) UpdateMongoDbUser(ctx context.Context, request openapi.UpdateMongoDbUserRequestObject) (openapi.UpdateMongoDbUserResponseObject, error) {
	user, err := s.core.MongoDbMgmt.UpdateMongoDbUser(ctx, request.Id, *request.Body, middleware.GetAuth(ctx))
	if err != nil {
		return nil, errors.Wrap(err, "failed to update MongoDB user")
	}
	return openapi.UpdateMongoDbUser200JSONResponse(user), nil
}

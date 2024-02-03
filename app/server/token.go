package server

import (
	"context"
	"github.com/kuzznya/letsdeploy/app/middleware"
	"github.com/kuzznya/letsdeploy/internal/openapi"
	"github.com/pkg/errors"
)

func (s Server) CreateTempToken(ctx context.Context, request openapi.CreateTempTokenRequestObject) (openapi.CreateTempTokenResponseObject, error) {
	token, err := s.core.Tokens.CreateTempToken(ctx, middleware.GetAuth(ctx))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create token")
	}
	return openapi.CreateTempToken200JSONResponse{Token: token}, nil
}

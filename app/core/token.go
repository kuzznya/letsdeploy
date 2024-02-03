package core

import (
	"context"
	"github.com/kuzznya/letsdeploy/app/apperrors"
	"github.com/kuzznya/letsdeploy/app/middleware"
	"github.com/redis/go-redis/v9"
	"math/rand"
	"time"
)

type Tokens interface {
	CreateTempToken(ctx context.Context, auth middleware.Authentication) (string, error)
}

type tokensImpl struct {
	rdb *redis.Client
}

var _ Tokens = (*tokensImpl)(nil)

func InitTokens(rdb *redis.Client) Tokens {
	return &tokensImpl{
		rdb: rdb,
	}
}

func (t tokensImpl) CreateTempToken(ctx context.Context, auth middleware.Authentication) (string, error) {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	tokenLen := 16
	b := make([]rune, tokenLen)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	token := string(b)

	err := t.rdb.Set(ctx, token, auth.Username, 1*time.Minute).Err()
	if err != nil {
		return "", apperrors.InternalServerErrorWrap(err, "failed to save token to Redis")
	}

	return token, nil
}

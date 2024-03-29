package middleware

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/kuzznya/letsdeploy/app/apperrors"
	"github.com/kuzznya/letsdeploy/internal/openapi"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"math/big"
	"net/http"
	"net/url"
	"time"
)

const authContextKey = "authentication"

type Authentication struct {
	Username string
	Token    string
}

var ServiceAccount = Authentication{
	Username: "letsdeploy-service-account",
	Token:    "TODO",
}

func CreateAuthMiddleware(cfg *viper.Viper) openapi.MiddlewareFunc {
	oidcProvider := cfg.GetString("oidc.provider")
	rsaKeys := getPublicKeys(oidcProvider)
	return func(c *gin.Context) {
		AuthMiddleware(c, cfg, rsaKeys)
	}
}

func AuthMiddleware(ctx *gin.Context, cfg *viper.Viper, rsaKeys map[string]*rsa.PublicKey) {
	oidcProvider := cfg.GetString("oidc.provider")
	headerValue := ctx.GetHeader("Authorization")
	if headerValue == "" {
		ctx.Next()
		return
	}
	if headerValue == "" || len(headerValue) < 8 || headerValue[:7] != "Bearer " {
		log.Debugf("Authorization header does not contain Bearer token")
		_ = ctx.Error(apperrors.Unauthorized("Authorization header does not contain Bearer token"))
		ctx.Abort()
		return
	}
	tokenString := headerValue[7:]

	token, err := parseToken(tokenString, rsaKeys)
	if err != nil {
		log.WithError(err).Errorln("Failed to parse JWT")
		_ = ctx.Error(apperrors.Forbidden("Failed to authenticate user"))
		ctx.Abort()
		return
	}
	if err := checkClaims(oidcProvider, token); err != nil {
		log.WithError(err).Errorf("Failed to check claims of token %+v", token)
		_ = ctx.Error(err)
		ctx.Abort()
		return
	}
	claim := cfg.GetString("oidc.username-claim")
	username := token.Claims.(jwt.MapClaims)[claim].(string)
	ctx.Set(authContextKey, &Authentication{Username: username, Token: tokenString})
	log.Debugf("User %s authenticated", username)

	ctx.Next()
}

// parseToken parses token and allows clock skew
func parseToken(tokenStr string, rsaKeys map[string]*rsa.PublicKey) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return rsaKeys[t.Header["kid"].(string)], nil
	})
	if ve, ok := err.(*jwt.ValidationError); ok && ve.Errors == jwt.ValidationErrorIssuedAt {
		delta := int(token.Claims.(jwt.MapClaims)["iat"].(float64) - float64(time.Now().Unix()))
		if delta < 3 {
			log.Warnf("Encountered lock skew %ds, skipping", delta)
			token.Valid = true
			return token, nil
		} else {
			return nil, errors.Wrapf(err, "time diff is %ds that is more than allowed", delta)
		}
	}
	return token, err
}

func GetAuth(ctx context.Context) Authentication {
	return *ctx.Value(authContextKey).(*Authentication)
}

func getPublicKeys(oidcProvider string) map[string]*rsa.PublicKey {
	jwksUri := getJwksUri(oidcProvider)

	response, err := http.Get(jwksUri)
	if err != nil {
		log.WithError(err).Panicln("Failed to fetch public keys for OIDC provider")
	}
	if response.StatusCode != http.StatusOK {
		log.WithError(err).Panicf(
			"Failed to fetch public keys for OIDC provider: server returned non-200 status: %d %s\n",
			response.StatusCode,
			response.Status)
	}

	var body map[string]any
	err = json.NewDecoder(response.Body).Decode(&body)
	if err != nil {
		log.WithError(err).Panicln("Failed to parse public keys response for OIDC provider")
	}

	rsaKeys := make(map[string]*rsa.PublicKey)
	for _, bodykey := range body["keys"].([]interface{}) {
		key := bodykey.(map[string]interface{})
		kid := key["kid"].(string)
		rsakey := new(rsa.PublicKey)
		number, _ := base64.RawURLEncoding.DecodeString(key["n"].(string))
		rsakey.N = new(big.Int).SetBytes(number)
		rsakey.E = 65537 // TODO take it from the "e" parameter in jwks_uri
		rsaKeys[kid] = rsakey
	}
	return rsaKeys
}

func getJwksUri(oidcProvider string) string {
	wellKnownUrl, err := url.JoinPath(oidcProvider, "/.well-known/openid-configuration")
	if err != nil {
		log.WithError(err).Panicln("Failed to create well-known config URL for OIDC provider")
	}
	response, err := http.Get(wellKnownUrl)
	if err != nil {
		log.WithError(err).Panicln("Failed to fetch well-known config for OIDC provider")
	}
	if response.StatusCode != http.StatusOK {
		log.WithError(err).Panicf(
			"Failed to fetch well-known config for OIDC provider: server returned non-200 status: %d %s\n",
			response.StatusCode,
			response.Status)
	}
	var wellKnownConfig map[string]any
	err = json.NewDecoder(response.Body).Decode(&wellKnownConfig)
	if err != nil {
		log.WithError(err).Panicln("Failed to parse well-known config for OIDC provider")
	}

	jwksUri := wellKnownConfig["jwks_uri"].(string)
	if jwksUri == "" {
		log.Panicln("Failed to get jwks_uri in well-known config for OIDC provider")
	}
	return jwksUri
}

func checkClaims(oidcProvider string, token *jwt.Token) error {
	if !token.Valid || token.Claims.(jwt.MapClaims)["iss"] != oidcProvider {
		return apperrors.Forbidden("Failed to authenticate user")
	}
	return nil
}

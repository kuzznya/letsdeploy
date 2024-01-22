package openapi

import _ "github.com/deepmap/oapi-codegen/v2/pkg/codegen"

//go:generate go run github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen -config oapi-codegen-models.yaml ../../api/openapi.yaml
//go:generate go run github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen -config oapi-codegen-server.yaml ../../api/openapi.yaml

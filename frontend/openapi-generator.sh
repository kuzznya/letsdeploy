#!/bin/bash

docker run --rm -v "${PWD}:/local" -v "${PWD}/../api/openapi.yaml:/api/openapi.yaml" \
    openapitools/openapi-generator-cli generate \
    -i /api/openapi.yaml \
    -g typescript-axios \
    -o /local/src/api/generated

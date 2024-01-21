FROM golang:1.21 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go generate github.com/kuzznya/letsdeploy/...
RUN go build github.com/kuzznya/letsdeploy/cmd/letsdeploy

FROM debian:bookworm-slim

RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY api ./api
COPY configs ./configs
COPY migrations ./migrations

COPY --from=build /app/letsdeploy .

EXPOSE 8080

ENTRYPOINT /app/letsdeploy

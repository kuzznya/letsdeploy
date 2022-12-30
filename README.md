# Letsdeploy

[![Backend CI/CD](https://github.com/kuzznya/letsdeploy/actions/workflows/backend.yml/badge.svg)](https://github.com/kuzznya/letsdeploy/actions/workflows/backend.yml)
[![Frontend CI/CD](https://github.com/kuzznya/letsdeploy/actions/workflows/frontend.yml/badge.svg)](https://github.com/kuzznya/letsdeploy/actions/workflows/frontend.yml)
[![Service status](https://shields.io/website?label=status&down_color=critical&down_message=down&up_color=success&up_message=up&url=https://letsdeploy.space)](https://letsdeploy.space)

Letsdeploy simplifies the process of your project deployment. 
It takes care of your services and the resources you depend on 
(called "managed service"): PostgreSQL, Redis, RabbitMQ, etc.

Letsdeploy works on top of Kubernetes cluster.

## Requirements

- Docker 20.0.0+
- minikube 1.2+
- go 1.15+

## Installation

1. Start minikube cluster:

    ```bash
    minikube start
    ```

2. Run `go generate`:

    ```bash
    go generate github.com/kuzznya/letsdeploy/...
    ```

3. Run openapi generator:

    - Windows:

        ```cmd
        .\frontend\openapi-generator.cmd
        ```
    - Linux/MacOS:
    
        ```bash
        ./frontend/openapi-generator.sh
        ```

4. Run `docker-compose`
    
    ```bash
    docker-compose up -d
    ```

5. Run server:

    ```bash
    go run .\cmd\letsdeploy\main.go
    ```

6. Run frontend:

    ```bash
    npm run dev --prefix frontend
    ```
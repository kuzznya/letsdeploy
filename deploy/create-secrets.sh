#!/bin/bash

kubectl --namespace junction create secret generic letsdeploy-secrets --from-env-file .env

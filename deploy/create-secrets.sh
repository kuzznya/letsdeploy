#!/bin/bash

kubectl --namespace letsdeploy create secret generic letsdeploy-secrets --from-env-file .env

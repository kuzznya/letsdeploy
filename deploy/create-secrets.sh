#!/bin/bash

set -e

kubectl get secret regcred --namespace default -o yaml | grep -v '^\s*namespace:\s' | kubectl apply --namespace letsdeploy -f -
kubectl --namespace letsdeploy create secret generic letsdeploy-secrets --from-env-file .env

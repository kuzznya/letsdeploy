#!/bin/bash

kubectl get secret regcred --namespace default -o yaml | kubectl apply --namespace letsdeploy -f -
kubectl --namespace letsdeploy create secret generic letsdeploy-secrets --from-env-file .env

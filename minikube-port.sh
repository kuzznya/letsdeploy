#!/bin/bash

set -e

docker inspect minikube | jq -r '.[].NetworkSettings.Ports."8443/tcp"[].HostPort'

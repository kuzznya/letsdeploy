#!/bin/bash

set -e

kubectl port-forward -n letsdeploy service/postgres 15432:5432

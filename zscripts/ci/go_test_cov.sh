#!/bin/bash

set -e

# Array of service directories
services=("dogalarm" "dogapm" "ordersvc" "skusvc" "usrsvc")

# Loop through each service and run tests
for service in "${services[@]}"; do
    cd "$service"
    go test -v -race -covermode=atomic -coverpkg=./... -coverprofile=coverage.out ./...
    go tool cover -func=coverage.out
    mv coverage.out ../coverage_${service}.out
    cd ..
done
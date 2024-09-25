#!/bin/bash

set -e

# Array of service directories
services=("dogalarm" "dogapm" "ordersvc" "skusvc" "usrsvc")

# Loop through each service and run tests
for service in "${services[@]}"; do
    cd "$service"
    go test -race ./...
    cd ..
done
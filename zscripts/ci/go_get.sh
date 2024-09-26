#!/bin/bash

set -e

# Array of service directories
services=("dogalarm" "dogapm" "ordersvc" "skusvc" "usrsvc" "protos")

# Loop through each service and run tests
for service in "${services[@]}"; do
    cd "$service"
    go get
    cd ..
done
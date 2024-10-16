#!/bin/bash

echo "Waiting for middlewares to be ready..."
sleep 3

echo "Starting applications..."
APP_NAME=skusvc /build/skusvc &
APP_NAME=usrsvc /build/usrsvc &
APP_NAME=ordersvc /build/ordersvc &

tail -f /dev/null

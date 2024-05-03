#!/bin/bash +x
GO111MODULE=off go build -o service-go && \
SERVICE_NAME=jwt-service-go \
GO_PORT=8000 \
GO_ENV=development \
GO_DEBUG=1 \
AWS_REGION=us-west-2 \
./service-go

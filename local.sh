#!/bin/bash +x
rm -rf go.mod go.sum &&
    GO111MODULE=on go mod init &&
    GO111MODULE=on go mod tidy &&
    GO111MODULE=on go get -d -u -v &&
    GO111MODULE=on go build -o service &&
    AWS_REGION=us-west-2 \
        GO_PORT=8000 \
        GO_DEBUG=1 \
        GO_ENV=development \
        ./service

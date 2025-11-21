#!/bin/bash

# Find where protoc-gen-validate is installed
PGV_PATH=$(go list -f '{{ .Dir }}' -m github.com/envoyproxy/protoc-gen-validate)

PATH=$PATH:$(go env GOPATH)/bin protoc \
    -I . \
    -I ${PGV_PATH} \
    --go_out=. --go_opt=paths=source_relative \
    --validate_out="lang=go:." --validate_opt=paths=source_relative \
    api/proto/v1/tasks.proto

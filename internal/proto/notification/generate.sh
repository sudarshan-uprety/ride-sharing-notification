#!/bin/bash

set -e

PROTO_DIR=.
PROTO_FILE=service.proto
OUTPUT_DIR=.

cd internal/proto/notification
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       service.proto

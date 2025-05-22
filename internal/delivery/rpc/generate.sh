#!/bin/bash

set -e

PROTO_DIR=.
PROTO_FILE=service.proto
OUTPUT_DIR=.

protoc --go_out=$OUTPUT_DIR --go_opt=paths=source_relative \
    --go-grpc_out=$OUTPUT_DIR --go-grpc_opt=paths=source_relative \
    $PROTO_DIR/$PROTO_FILE

echo "Generated gRPC code from $PROTO_FILE"

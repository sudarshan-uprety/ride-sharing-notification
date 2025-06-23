#!/bin/bash

set -e

# Navigate to the shared protos directory
cd internal/proto/shared/api/notifications

# Run the generation script from the shared repo
./generate.sh

# Copy or move the generated files to where your service expects them
cp -r grpcFiles/* ../../../notification/

#!/usr/bin/bash

echo "Hello 2"

export GIN_MODE=debug
export REVERSE_PROXY_HOME="/opt/source/templates/*"
export IN_CLUSTER="true"
air
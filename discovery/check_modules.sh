#!/bin/bash

set -e

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
MODULES=(
  "$ROOT/discovery"
  "$ROOT/discovery/consul"
  "$ROOT/discovery/etcd"
  "$ROOT/discovery/nacos"
  "$ROOT/discovery/redis"
  "$ROOT/discovery/zookeeper"
  "$ROOT/discovery/k8s"
)

echo "Running discovery unit tests..."
for dir in "${MODULES[@]}"; do
  echo "\n==> go test ./... in $dir"
  (cd "$dir" && go test ./...)
done

if [ "${RUN_INTEGRATION:-}" = "1" ]; then
  echo "\nRunning discovery integration tests..."
  for dir in "${MODULES[@]}"; do
    echo "\n==> go test -tags=integration ./... in $dir"
    (cd "$dir" && go test -tags=integration ./...)
  done
else
  echo "\nSkip integration tests (set RUN_INTEGRATION=1 to enable)."
fi

echo "\nModule checks completed."

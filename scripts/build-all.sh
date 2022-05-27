#!/bin/bash
set -e

COMMIT_HASH=$(git rev-parse HEAD)
COMMIT_HASH=${COMMIT_HASH:0:8}
VERSION=$(date -u +%Y%m%d)_${COMMIT_HASH}

ROOT_DIR=$(cd $(dirname $(dirname $0)) && pwd)
#echo "ROOT_DIR=$ROOT_DIR"

cd $ROOT_DIR
IMPORT_PATH=$(go list)/cmd
#echo $IMPORT_PATH

rm -rf $ROOT_DIR/out
mkdir $ROOT_DIR/out

echo "building $VERSION"

CGO_ENABLED=0 GOARCH=amd64 GOOS=windows go build -tags release -ldflags "-s -X $IMPORT_PATH.version=$VERSION" -o $ROOT_DIR/out/yabrc-win-amd64.exe main.go
CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -tags release -ldflags "-s -X $IMPORT_PATH.version=$VERSION" -o $ROOT_DIR/out/yabrc-linux-amd64 main.go
CGO_ENABLED=0 GOARCH=amd64 GOOS=darwin go build -tags release -ldflags "-s -X $IMPORT_PATH.version=$VERSION" -o $ROOT_DIR/out/yabrc-darwin-amd64 main.go
CGO_ENABLED=0 GOARCH=arm64 GOOS=darwin go build -tags release -ldflags "-s -X $IMPORT_PATH.version=$VERSION" -o $ROOT_DIR/out/yabrc-darwin-arm64 main.go

ls -lh out | awk 'NR > 1 { print $9 " " $5 }'
echo "build complete"
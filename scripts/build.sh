#!/bin/sh

VERSION=${1:-dev}
GOOS=${2:-linux}
DOCKER_REPO="bartlettc/kubeviz-agent"

# Directory to house our binaries
mkdir -p bin

# Build the binary in Docker and extract it from the container
docker build --build-arg VERSION=${VERSION} --build-arg GOOS=${GOOS} -t ${DOCKER_REPO}:${VERSION}-${GOOS} ./
docker run --rm --name kubeviz-agent-build -d --entrypoint "" ${DOCKER_REPO}:${VERSION}-${GOOS} sh -c "sleep 30"
docker cp kubeviz-agent-build:/usr/bin/kubeviz-agent bin
docker stop kubeviz-agent-build

# Zip up the binary
cd bin
zip kubeviz-agent-${GOOS}-${VERSION}.zip kubeviz-agent
rm kubeviz-agent

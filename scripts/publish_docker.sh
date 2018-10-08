#!/bin/sh

SOURCE_VERSION=${1}
PUBLISH_VERSION=${SOURCE_VERSION}
LATEST=${2:-false}
DOCKER_REPO="bartlettc/kubeviz-agent"

docker tag ${DOCKER_REPO}:${SOURCE_VERSION}-linux ${DOCKER_REPO}:${PUBLISH_VERSION}
echo "${DOCKER_PASSWORD}" | docker login -u "${DOCKER_USERNAME}" --password-stdin
docker push ${DOCKER_REPO}:${PUBLISH_VERSION}

if [ ${LATEST} == "true" ]; then

  docker tag ${DOCKER_REPO}:${SOURCE_VERSION}-linux ${DOCKER_REPO}:latest
  echo "${DOCKER_PASSWORD}" | docker login -u "${DOCKER_USERNAME}" --password-stdin
  docker push ${DOCKER_REPO}:latest

fi

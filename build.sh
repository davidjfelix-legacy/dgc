#!/bin/bash

set -o nounset
set -o errexit

docker build -t hatchery/dgc-build .
docker run -v /var/run/docker.sock:/var/run/docker.sock \
  -v $(which docker):$(which docker) \
  -it --entrypoint=docker \
  --name dgc-build hatchery/dgc-build \
  build -t hatcher/dgc /opt/dgc

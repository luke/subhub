#!/bin/bash
docker build -t screencloud/subhub_build .
docker run -v /var/run/docker.sock:/var/run/docker.sock -v $(which docker):$(which docker) -ti --name subhub_build screencloud/subhub_build
docker rm subhub_build
docker rmi screencloud/subhub_build
# [Dockerized] (http://www.docker.com) [subhub](https://registry.hub.docker.com/u/screencloud/subhub/)

A docker image for subhub. This is created as a single static executable, so there are multiple passes through Docker to first build the static executable and then to package it under an empty (scratch) base image.

Credit: This recipe for dockerizing a go app was copied from GNATSD repo

Updating the image run.. 

./build.sh
docker login
docker push screencloud/subhub
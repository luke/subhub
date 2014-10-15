FROM google/golang:1.3

MAINTAINER screencloud

RUN CGO_ENABLED=0 go get -a -ldflags '-s'  github.com/screencloud/subhub/cmd/subhub
COPY Dockerfile.final /gopath/bin/Dockerfile

CMD docker build -t screencloud/subhub /gopath/bin

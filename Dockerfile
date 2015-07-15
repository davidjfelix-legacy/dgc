FROM google/golang

ENV CGO_ENABLED 0

COPY ./dgc /opt/dgc
WORKDIR /opt/dgc

RUN go get github.com/codegangsta/cli
RUN go get github.com/fsouza/go-dockerclient

RUN go build -a -ldflags '-s'
RUN ln -s /opt/dgc/dgc /usr/local/bin

ENTRYPOINT ["dgc"]

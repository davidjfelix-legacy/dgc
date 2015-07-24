FROM google/golang

ENV CGO_ENABLED 0


RUN go get github.com/codegangsta/cli
RUN go get github.com/fsouza/go-dockerclient

COPY ./dgc /opt/dgc
WORKDIR /opt/dgc

RUN go build -a -ldflags '-s'
RUN ln -s /opt/dgc/dgc /usr/local/bin

ENTRYPOINT ["dgc"]

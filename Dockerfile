FROM google/golang

ENV CGO_ENABLED 0

COPY ./dgc /opt/dgc
WORKDIR /opt/dgc

RUN go build -a -ldflags '-s'
RUN ln -s /opt/dgc/dgc /usr/local/bin

ENTRYPOINT ["dgc"]

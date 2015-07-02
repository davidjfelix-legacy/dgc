FROM google/golang

ENV CGO_ENABLED 0

COPY ./dgc /opt/dgc
WORKDIR /opt/dgc

RUN go build -a -ldflags '-s'
RUN ln -s /opt/dgc/dgc /usr/local/bin

CMD ["dgc"]
#CMD ["docker", "build", "-t", "hatchery/dgc", "/opt/dgc"] 

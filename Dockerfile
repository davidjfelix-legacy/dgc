FROM scratch
ADD dgc/dgc /dgc
ENTRYPOINT ["/dgc"]

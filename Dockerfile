FROM golang

ENV GIN_MODE=release

WORKDIR /opt
COPY goflow-example goflow-example
COPY assets assets

EXPOSE 8181

CMD ["./goflow-example"]

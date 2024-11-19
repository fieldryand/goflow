FROM centos

ENV GIN_MODE=release

WORKDIR /opt
COPY goflow-example-simple goflow-example-simple
COPY ui ui

EXPOSE 8181

CMD ["./goflow-example-simple"]

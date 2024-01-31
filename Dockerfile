FROM centos

ENV GIN_MODE=release

WORKDIR /opt
COPY goflow-example goflow-example
COPY ui ui

EXPOSE 8181

CMD ["./goflow-example"]

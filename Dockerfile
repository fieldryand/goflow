FROM centos

ENV GIN_MODE=release

WORKDIR /opt
COPY goflow-example goflow-example

EXPOSE 8181

CMD ["./goflow-example"]

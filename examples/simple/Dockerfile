FROM golang:1.14

ENV GIN_MODE=release

RUN apt-get update -yq \
    && apt-get -yq install curl \
    && curl -L https://deb.nodesource.com/setup_12.x | bash \
    && apt-get update -yq \
    && apt-get install -yq \
        nodejs

WORKDIR /go/src/app
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

RUN $GOPATH/src/github.com/fieldryand/goflow/compile_assets.sh

CMD ["app"]

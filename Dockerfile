FROM golang:1.16

ENV GIN_MODE=release

RUN apt-get update -yq \
    && apt-get -yq install curl git \
    && curl -L https://deb.nodesource.com/setup_12.x | bash \
    && apt-get update -yq \
    && apt-get install -yq nodejs \
    && cd $GOPATH/src \
    && mkdir -p github.com/fieldryand \
    && git clone https://github.com/fieldryand/goflow.git github.com/fieldryand/goflow \
    && $GOPATH/src/github.com/fieldryand/goflow/compile_assets.sh

WORKDIR /go/src/app
COPY . .

RUN go get -d -v ./... \
    && go install -v ./...

EXPOSE 8181

CMD ["example"]

FROM golang:1.14

ENV GIN_MODE=release

WORKDIR /go/src/app
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["app"]

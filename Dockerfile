ARG VERSION=latest
FROM golang:${VERSION}

RUN apt-get update && apt-get install -y openjdk-17-jre file protobuf-compiler
COPY . /go/src/github.com/pact-foundation/pact-go

WORKDIR /go/src/github.com/pact-foundation/pact-go

CMD ["make", "test"]
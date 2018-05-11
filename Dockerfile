FROM golang:1.10.1

# Install pact ruby standalone binaries
RUN curl -LO https://github.com/pact-foundation/pact-ruby-standalone/releases/download/v1.37.0/pact-1.37.0-linux-x86_64.tar.gz; \
    tar -C /usr/local -xzf pact-1.37.0-linux-x86_64.tar.gz; \
    rm pact-1.37.0-linux-x86_64.tar.gz

ENV PATH /usr/local/pact/bin:$PATH

COPY . /go/src/github.com/pact-foundation/pact-go

WORKDIR /go/src/github.com/pact-foundation/pact-go

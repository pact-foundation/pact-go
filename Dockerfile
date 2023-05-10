FROM golang:1.18

# Install pact ruby standalone binaries
RUN curl -LO https://github.com/you54f/pact-ruby-standalone/releases/download/v2.2.1/pact-2.2.1-linux-x86_64.tar.gz; \
    tar -C /usr/local -xzf pact-2.2.1-linux-x86_64.tar.gz; \
    rm pact-2.2.1-linux-x86_64.tar.gz

ENV PATH /usr/local/pact/bin:$PATH

COPY . /go/src/github.com/pact-foundation/pact-go

WORKDIR /go/src/github.com/pact-foundation/pact-go

FROM golang:latest

COPY . /lldescript
WORKDIR /lldescript

RUN go install github.com/ramya-rao-a/go-outline@latest && \
    go install honnef.co/go/tools/cmd/staticcheck@latest

# Stage 1: Build the custom Go runtime
FROM debian:bookworm-slim AS goruntime-builder

RUN apt-get update && apt-get install -y bash build-essential curl git

# Install bootstrap Go compiler for building custom runtime
ENV GOROOT_BOOTSTRAP=/usr/local/go

RUN curl -fsSL https://go.dev/dl/go1.24.1.linux-amd64.tar.gz | tar -C /usr/local -xzf -

COPY go-patch /GoCR/go-patch
WORKDIR /GoCR/go-patch/src
RUN bash make.bash

# Stage 2: Build the Go app using the standard Go runtime
FROM golang:1.24 AS app-builder

WORKDIR /GoCR/advocate
COPY goCr /GoCR/goCr
COPY go-patch /GoCr/go-patch

RUN go build -o app .

# Stage 3: Final runtime image
FROM debian:bookworm-slim

WORKDIR /GoCR

COPY --from=goruntime-builder /GoCR/go-patch ./go-patch
COPY --from=app-builder /GoCR/goCR/app ./goCR/app

WORKDIR /prog

WORKDIR /GoCR/goCR

ENTRYPOINT ["./app"]
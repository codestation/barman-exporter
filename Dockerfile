FROM golang:1.17 as builder

ARG CI_COMMIT_TAG
ARG CI_COMMIT_BRANCH
ARG CI_COMMIT_SHA
ARG CI_PIPELINE_CREATED_AT
ARG GOPROXY
ENV GOPROXY=${GOPROXY}

WORKDIR /src
COPY go.mod go.sum /src/
RUN go mod download
COPY . /src/

RUN set -ex; \
    CGO_ENABLED=0 go build -o release/barman-exporter \
    -ldflags "-w -s \
    -X main.Version=${CI_COMMIT_TAG:-$CI_COMMIT_BRANCH} \
    -X main.Commit=$(echo "$CI_COMMIT_SHA" | cut -c1-8) \
    -X main.BuildTime=${CI_PIPELINE_CREATED_AT}"

FROM debian:bullseye
LABEL maintainer="codestation <codestation404@gmail.com>"

RUN set -ex; \
    apt-get update; \
    DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends ca-certificates; \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /src/release/barman-exporter /bin/barman-exporter

ENTRYPOINT ["/bin/barman-exporter"]

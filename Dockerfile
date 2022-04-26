FROM golang:1.17-alpine AS build-env

ARG VERSION
ARG COMMIT_HASH
ARG CI

ENV GOFLAGS="-mod=vendor"
ENV CGO_ENABLED=0

WORKDIR /build
ADD . /build

RUN apk add --update --no-cache build-base make git libc-dev

RUN \
    if [ -z "$CI" ] ; then \
        echo "runs outside of CI"; \
        VERSION=$(git rev-parse --abbrev-ref HEAD); \
        COMMIT_HASH=$(git rev-parse --short HEAD); \
    fi && \
    DATE=$(date +%FT%T%z); \
    LDFLAGS="-s -w -X main.Version=${VERSION} -X main.CommitHash=${COMMIT_HASH} -X main.CompileDate=${DATE}"; \
    go build -o /go/bin/hooks -ldflags "${LDFLAGS}" && \
    /go/bin/hooks --version


FROM alpine:3.15

WORKDIR /app
COPY --from=build-env /go/bin/hooks /app
COPY --from=build-env /build/static/ /app/static/
COPY --from=build-env /build/templates/ /app/templates/

ENV HOOKS_HOST "0.0.0.0"
ENV HOOKS_PORT "8080"
ENV HOOKS_URL "http://0.0.0.0:8080"
ENV BOLT_PATH "/app/var"

EXPOSE ${HOOKS_PORT}

ENTRYPOINT ["/app/hooks"]

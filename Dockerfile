# Stage 1 - Binary Build
# BUILD_X args should be passed at build time as docker build args
FROM golang:1.19.0-alpine3.15 AS builder
ARG BIN_VERSION
ARG GIT_COMMIT
ARG GIT_SHA
ARG GIT_TAG
ARG GIT_DIRTY
ENV BIN_OUTDIR=./
ENV BIN_NAME=podchaosmonkey
RUN apk update && apk add build-base git libressl-dev
WORKDIR /usr/src/app
# install dependencies in separate docker layer
COPY go.mod .
COPY go.sum .
RUN go mod download
# copy application source and build
COPY ./ .
RUN make build

# Stage 2 - Final Image
# The application should be statically linked
FROM alpine:3.10
RUN apk update \
	&& apk add --no-cache ca-certificates \
	&& rm -rf /var/cache/apk/* \
    && addgroup podchaosmonkey \
	&& adduser -D -H -G podchaosmonkey podchaosmonkey
COPY --from=builder /usr/src/app/podchaosmonkey /usr/bin/podchaosmonkey
ENTRYPOINT ["podchaosmonkey"]
USER podchaosmonkey
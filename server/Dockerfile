# ----------------------------------------------------
# Base
# ----------------------------------------------------
FROM golang:1.14.2-alpine3.11 AS build_base
RUN apk update && \
    apk add --no-cache \
    bash \
    build-base \
    tzdata \
    git \
    sqlite-dev \
    sqlite
RUN mkdir -p {/app/learnalist-api,/app/bin}
WORKDIR /app/learnalist-api/server

COPY go.mod .
COPY go.sum .
RUN go mod download

# ----------------------------------------------------
# Build + Test
# ----------------------------------------------------
ARG GIT_COMMIT
ENV GIT_COMMIT ${GIT_COMMIT}
ARG GIT_HASH_DATE
ENV GIT_HASH_DATE ${GIT_HASH_DATE}

FROM build_base AS build
WORKDIR /app/learnalist-api/server
COPY --from=build_base /app/learnalist-api/server .
COPY . .
#RUN go test --tags="libsqlite3 json1" `go list ./... | grep -vE 'event|integrations|e2e'` -covermode=count -coverprofile=profile.cov

ENV GOOS linux
ENV GOARCH amd64
ENV CGO_ENABLED 1
RUN './build.sh'

# ----------------------------------------------------
# Release
# ----------------------------------------------------
FROM learnalist-prod-base:latest AS release

RUN mkdir -p {/app/learnalist-api,/app/bin}
COPY --from=build /app/learnalist-api/server/learnalist-cli /app/bin/learnalist-cli

WORKDIR /app
ENTRYPOINT ["/app/bin/learnalist-cli"]

EXPOSE 8000

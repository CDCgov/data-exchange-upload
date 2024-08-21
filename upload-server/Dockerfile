FROM docker.io/library/golang:1.22.1-alpine AS builder

ARG REPO="cdcgov/data-exchange-upload"
ARG LATEST_RELEASE_VERSION="unspecified"
ARG GIT_SHORT_SHA="unspecified"

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -ldflags "-X github.com/cdcgov/data-exchange-upload/upload-server/internal/version.GitRepo=${REPO}  \
    -X github.com/cdcgov/data-exchange-upload/upload-server/internal/version.LatestReleaseVersion=${LATEST_RELEASE_VERSION} \
    -X github.com/cdcgov/data-exchange-upload/upload-server/internal/version.GitShortSha=${GIT_SHORT_SHA}" -o ./dextusd ./cmd/main.go

FROM docker.io/library/alpine:3.20.2

WORKDIR /app

COPY --from=builder /app/dextusd .

EXPOSE 8080

CMD ["./dextusd"]

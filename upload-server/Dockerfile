FROM docker.io/library/golang:1.22.0-alpine AS builder

WORKDIR /app 

COPY go.mod go.sum ./ 

RUN go mod download

COPY . .

RUN go build -o ./dextusd ./cmd/main.go

#

FROM alpine:latest 

WORKDIR /app 

COPY --from=builder /app/dextusd .

EXPOSE 8080

CMD ["./dextusd"]

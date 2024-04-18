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
COPY --from=builder /app/configs ./configs  
COPY --from=builder /app/uploads ./uploads 

EXPOSE 8080

CMD ["./dextusd", "-env", "local_to_azure"]

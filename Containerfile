FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o sabeel ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -o zitadel-bootstrap ./cmd/zitadel-bootstrap
RUN CGO_ENABLED=0 GOOS=linux go build -o sabeel-migrate ./cmd/migrate

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app
COPY --from=builder /app/sabeel .
COPY --from=builder /app/zitadel-bootstrap .
COPY --from=builder /app/sabeel-migrate .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["./sabeel"]

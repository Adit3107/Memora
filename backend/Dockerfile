FROM golang:1.27-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/mindshelf-backend ./cmd/server

FROM alpine:3.22
WORKDIR /app

RUN apk add --no-cache ca-certificates wget

COPY --from=builder /out/mindshelf-backend /app/mindshelf-backend

EXPOSE 8080

CMD ["/app/mindshelf-backend"]

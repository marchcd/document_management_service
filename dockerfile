FROM golang:1.26-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main ./cmd/main.go

FROM alpine:latest
WORKDIR /root/
RUN apk --no-cache add ca-certificates

COPY --from=builder /app/main .
COPY --from=builder /app/static ./static
COPY --from=builder /app/migrations ./migrations

RUN chmod +x ./main

CMD ["./main"]
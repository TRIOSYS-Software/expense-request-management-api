FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY . .

RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/main.go

FROM alpine:latest

WORKDIR /app

RUN mkdir /app/uploads

COPY fcm_credentials.json .

COPY --from=builder /app/main .

CMD ["/app/main"]
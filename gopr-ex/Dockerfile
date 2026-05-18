FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o blog-api ./main.go

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/blog-api ./blog-api
COPY .env.example ./.env.example

ENV SERVER_HOST=0.0.0.0
ENV SERVER_PORT=8080
ENV TOKEN_SECRET=change-me-in-production
ENV DATA_DIR=/app/data
ENV LOG_FILE=/app/data/log.txt

RUN mkdir -p /app/data

EXPOSE 8080

CMD ["./blog-api"]

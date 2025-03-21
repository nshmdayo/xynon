# ステージ１
FROM golang:1.23-alpine3.20 AS go
WORKDIR /
COPY go.mod ./
RUN go mod download
WORKDIR /app
COPY app/ ./
RUN go build -o main /app/main.go

# ステージ２
FROM alpine:3.20
WORKDIR /app
COPY --from=go /app/main .
USER 1001
CMD [ "/app/main" ]
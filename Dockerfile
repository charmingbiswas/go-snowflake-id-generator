FROM golang:latest AS builder
WORKDIR /app
COPY . .
RUN GOOS=linux go build -o snowflake ./cmd/main.go
EXPOSE 8000
ENTRYPOINT [ "./snowflake" ]
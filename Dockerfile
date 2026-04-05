# Dockerfile
FROM golang:latest AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o snowflake ./cmd/main.go

FROM scratch                       
COPY --from=builder /app/snowflake /snowflake
EXPOSE 8080
ENTRYPOINT ["/snowflake"]
FROM golang:latest AS builder
WORKDIR /app
COPY . .
RUN GOOS=linux go build -o snowflake ./cmd/main.go

FROM scratch
COPY --from=builder /app/snowflake /snowflake
EXPOSE 8000
ENTRYPOINT [ "./snowflake" ]
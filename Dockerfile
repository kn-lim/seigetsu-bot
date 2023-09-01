# Builder
FROM golang:1.21 as builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o seigetsu-bot .

# Runner
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/seigetsu-bot .
ENTRYPOINT ["./seigetsu-bot"]
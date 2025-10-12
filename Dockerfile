FROM golang:1.23 AS builder
WORKDIR /app
COPY coroot/ .
RUN go mod tidy
RUN go build -o rca-backend

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/rca-backend .
CMD ["./rca-backend"]
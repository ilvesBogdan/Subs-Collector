FROM golang:1.25 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/app ./cmd/app

FROM alpine:3.20
RUN adduser -D -u 10001 appuser
WORKDIR /
COPY --from=builder /bin/app /app
ENV PORT=8080
EXPOSE 8080
USER appuser
ENTRYPOINT ["/app"]



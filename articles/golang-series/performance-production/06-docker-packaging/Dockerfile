FROM golang:1.24 AS builder
WORKDIR /src
COPY go.mod .
COPY main.go .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /app .

FROM scratch
COPY --from=builder /app /app
ENTRYPOINT ["/app"]

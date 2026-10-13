# ^^^^^^^^^^^^^^^ BUILD ^^^^^^^^^^^^^^^
FROM golang:1.27-alpine3.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /trip-service ./cmd/trip-service

# ^^^^^^^^^^^^^^^ RUNTIME ^^^^^^^^^^^^^^^
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /trip-service /trip-service

EXPOSE 8080

ENTRYPOINT ["/trip-service"]

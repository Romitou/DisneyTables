# Build Stage
FROM golang:1.20-alpine AS builder

WORKDIR /app
RUN apk add --no-cache git ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o disneytables main.go

# Runtime Stage
FROM alpine:3.17

WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/disneytables /app/disneytables

EXPOSE 8080
ENTRYPOINT ["/app/disneytables"]

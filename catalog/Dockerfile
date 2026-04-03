FROM golang:1.24-alpine AS builder

WORKDIR /workspace

# Copy workspace files
COPY go.work go.work.sum ./
COPY catalog/ catalog/
COPY common/ common/

WORKDIR /workspace/catalog
# Download dependencies
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/main.go

# Minimal runtime image
FROM gcr.io/distroless/static-debian12

WORKDIR /

COPY --from=builder /workspace/catalog/main .

EXPOSE 8080 9090

ENTRYPOINT ["/main"]

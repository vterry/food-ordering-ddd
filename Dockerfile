FROM golang:alpine AS builder

WORKDIR /app

# Copy the entire workspace
COPY . .

# Enable Go workspace support explicitly
ENV GOWORK=/app/go.work

# Build the specified service
ARG SERVICE_PATH
RUN go build -o /app/service $SERVICE_PATH/cmd/server/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/service .

# Default envs for local dev
ENV MYSQL_DSN="root:root@tcp(mysql:3306)/food_project?parseTime=true"
ENV RABBITMQ_URL="amqp://guest:guest@rabbitmq:5672/"

CMD ["./service"]

FROM golang:1.23.2-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -trimpath -o ai-connect

FROM alpine:latest

LABEL org.opencontainers.image.source="https://github.com/dhbin/ai-connect"
LABEL org.opencontainers.image.description="ai-connect"

WORKDIR /root/

COPY --from=builder /app/ai-connect .

# Expose port 9090
EXPOSE 9090

# Command to run the application
ENTRYPOINT ["./ai-connect"]
# Build stage
FROM golang:1.19-alpine AS builder

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binaries
RUN cd bin/agent && go build -o agent
RUN cd bin/rendezvous && go build -o rendezvous
RUN cd bin/client && go build -o client

# Runtime stage - use Alpine 3.16 for better compatibility
FROM alpine:3.16

# Set working directory
WORKDIR /app

# Copy binaries from builder
COPY --from=builder /build/bin/agent/agent /app/bin/agent/agent
COPY --from=builder /build/bin/rendezvous/rendezvous /app/bin/rendezvous/rendezvous
COPY --from=builder /build/bin/client/client /app/bin/client/client

# Copy scripts
COPY bin/test.sh /app/bin/test.sh
RUN chmod +x /app/bin/test.sh

# Create logs directory
RUN mkdir -p /app/bin/logs

# Set working directory to bin
WORKDIR /app/bin

# Default command
CMD ["/bin/sh"]


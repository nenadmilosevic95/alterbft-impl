# Build stage
FROM golang:1.18-alpine AS builder

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

# Install bash for the test script
RUN apk add --no-cache bash || true

# Set working directory
WORKDIR /app/bin

# Create directory structure
RUN mkdir -p agent rendezvous client logs

# Copy binaries from builder
COPY --from=builder /build/bin/agent/agent ./agent/agent
COPY --from=builder /build/bin/rendezvous/rendezvous ./rendezvous/rendezvous
COPY --from=builder /build/bin/client/client ./client/client

# Copy scripts
COPY --from=builder /build/bin/test.sh ./test.sh
COPY --from=builder /build/bin/xcompile.sh ./xcompile.sh
COPY --from=builder /build/bin/xclean.sh ./xclean.sh

# Make scripts executable
RUN chmod +x test.sh xcompile.sh xclean.sh 2>/dev/null || true

# Default command
CMD ["/bin/sh"]


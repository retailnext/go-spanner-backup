FROM golang:1.20.5-alpine3.18 as builder

# Create and change to the app directory.
WORKDIR /app

# Retrieve application dependencies.
# This allows the container build to reuse cached dependencies.
# Expecting to copy go.mod and if present go.sum.
COPY go.* ./
RUN go mod download

# Copy local code to the container image.
COPY . ./

# Build the binary.
RUN go build -v -o spannerbackup

FROM alpine:3.18
# RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
#     ca-certificates && \
#     rm -rf /var/lib/apt/lists/*

# Copy the binary to the production image from the builder stage.
COPY --from=builder /app/spannerbackup /spannerbackup

# Entrypoint
Entrypoint ["/spannerbackup"]

# Run the web service by default on container startup.
CMD ["service"]
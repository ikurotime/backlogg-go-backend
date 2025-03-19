# syntax=docker/dockerfile:1

# Comments are provided throughout this file to help you get started.
# If you need more help, visit the Dockerfile reference guide at
# https://docs.docker.com/go/dockerfile-reference/

# Want to help us make this template better? Share your feedback here: https://forms.gle/ybq9Krt8jtBL3iCk7

################################################################################
# Create a stage for building the application.
FROM golang:1.23.3-alpine AS build

WORKDIR /app

# Install necessary build tools
RUN apk add --no-cache git

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Set build environment
ARG APP_ENV=production
ENV APP_ENV=${APP_ENV}

# Create necessary directories
RUN mkdir -p config

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd

################################################################################
# Create a new stage for running the application that contains the minimal
# runtime dependencies for the application. This often uses a different base
# image from the build stage where the necessary files are copied from the build
# stage.
#
# The example below uses the alpine image as the foundation for running the app.
# By specifying the "latest" tag, it will also use whatever happens to be the
# most recent version of that image when you build your Dockerfile. If
# reproducibility is important, consider using a versioned tag
# (e.g., alpine:3.17.2) or SHA (e.g., alpine@sha256:c41ab5c992deb4fe7e5da09f67a8804a46bd0592bfdf0b1847dde0e0889d2bff).
FROM alpine:latest

# Install CA certificates and YAML tools
RUN apk --no-cache add ca-certificates tzdata yq

# Create a non-root user
RUN adduser -D -g '' appuser

WORKDIR /app

# Copy built executable
COPY --from=build /app/server .

# Copy config directory
COPY --from=build /app/config /app/config

# Create entrypoint script with YAML validation
RUN printf '#!/bin/sh\n\
ENV_FILE="/app/config/.env.$APP_ENV"\n\
if [ ! -f "$ENV_FILE" ]; then\n\
  echo "ERROR: Environment file $ENV_FILE not found"\n\
  exit 1\n\
fi\n\
# Validate YAML format\n\
if ! cat "$ENV_FILE" | yq > /dev/null 2>&1; then\n\
  echo "ERROR: Invalid YAML format in $ENV_FILE"\n\
  echo "Content of $ENV_FILE:"\n\
  cat "$ENV_FILE"\n\
  exit 1\n\
fi\n\
echo "Using environment configuration: $ENV_FILE"\n\
exec /app/server "$@"\n' > /app/entrypoint.sh && \
    chmod +x /app/entrypoint.sh

# Set ownership
RUN chown -R appuser:appuser /app

# Set runtime environment
ARG APP_ENV=production
ENV APP_ENV=${APP_ENV}

USER appuser

# Expose the port that the application listens on.
EXPOSE 8080

# What the container should run when it is started.
ENTRYPOINT ["/app/entrypoint.sh"]

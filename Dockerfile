# FROM golang:1.23.4 as build
# WORKDIR /app
# COPY . .
# RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o cloudrun

# FROM scratch
# WORKDIR /app
# COPY --from=build /app/cloudrun .
# ENTRYPOINT ["./cloudrun"]

# Stage 1: Build the Go application
FROM golang:1.23.4 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the go.mod and go.sum files to download dependencies
# This is a key step for optimizing Docker caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the application
# CGO_ENABLED=0 creates a statically linked binary, which is crucial for the 'scratch' base image
# GOOS=linux ensures the binary is built for the target operating system (Cloud Run/Linux)
# GOARCH=amd64 is the architecture for Cloud Run
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o cloudrun .

# -----------------------------------------------------------------------------------------

# Stage 2: Create a minimal final image from scratch
FROM scratch

# Set the working directory
WORKDIR /app

# Copy the built binary from the builder stage
# The 'scratch' base image is tiny and secure, containing only the binary.
COPY --from=builder /app/cloudrun .

# Copy any necessary certificates if your application makes HTTPS requests to external services
# If you make HTTPS calls, uncomment the following line:
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Define the entrypoint to run the built application
# The port is automatically handled by the Cloud Run environment.
ENTRYPOINT ["./cloudrun"]
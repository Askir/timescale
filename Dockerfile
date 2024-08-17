# Start from the official Go image
FROM golang:1.23-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files
COPY . ./

# Build the application
RUN go build -o benchmark

# Start a new stage from scratch
FROM alpine:latest  

# Copy the binary from the builder stage
COPY --from=builder /app/benchmark /benchmark

COPY cpu_usage.csv /cpu_usage.csv

COPY query_params.csv /query_params.csv

# Set the binary as the entrypoint of the container
ENTRYPOINT ["/benchmark"]
